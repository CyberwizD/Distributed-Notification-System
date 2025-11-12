import asyncio
import json
import logging
import time
from typing import Any, Dict

import jinja2
import pika
from pika.adapters.blocking_connection import BlockingChannel

from app.clients.template_client import TemplateClient
from app.config.settings import settings
from app.email_sender import EmailSender
from app.services.status_store import StatusStore

logger = logging.getLogger(__name__)


class BaseConsumer:
    def __init__(self) -> None:
        self.connection: pika.BlockingConnection | None = None
        self.channel: BlockingChannel | None = None

    def connect(self) -> bool:
        try:
            self.connection = pika.BlockingConnection(pika.URLParameters(settings.rabbitmq_url))
            self.channel = self.connection.channel()
            self.channel.queue_declare(
                queue=settings.email_queue,
                durable=True,
                arguments={
                    "x-dead-letter-exchange": "",
                    "x-dead-letter-routing-key": settings.failed_queue,
                },
            )
            self.channel.queue_declare(queue=settings.failed_queue, durable=True)
            return True
        except Exception as exc:
            logger.error("Failed to connect to RabbitMQ: %s", exc)
            return False

    def stop(self) -> None:
        if self.connection and not self.connection.is_closed:
            self.connection.close()


class EmailConsumer(BaseConsumer):
    def __init__(self) -> None:
        super().__init__()
        self.email_sender = EmailSender()
        self.template_client = TemplateClient()
        self.status_store = StatusStore()
        self._jinja_env = jinja2.Environment(autoescape=True)

    def _render(self, template: str, variables: Dict[str, Any]) -> str:
        try:
            return self._jinja_env.from_string(template).render(**variables)
        except Exception as exc:
            logger.warning("Template rendering failed (%s); returning raw template", exc)
            return template

    def start_consuming(self) -> bool:
        backoff = 1.0
        for attempt in range(1, 6):
            if self.connect():
                break
            logger.warning("RabbitMQ connection attempt %d failed", attempt)
            time.sleep(backoff)
            backoff = min(backoff * 2, 10.0)
        else:
            return False

        assert self.channel is not None
        self.channel.basic_qos(prefetch_count=1)
        self.channel.basic_consume(
            queue=settings.email_queue,
            on_message_callback=self.process_message,
        )
        logger.info("Email consumer started on queue %s", settings.email_queue)
        self.channel.start_consuming()
        return True

    def process_message(self, ch: BlockingChannel, method, properties, body: bytes) -> None:
        try:
            envelope = json.loads(body)
            logger.debug("Processing envelope: %s", envelope)
            request_id = envelope.get("request_id")
            user = envelope.get("user") or {}
            template_info = envelope.get("template") or {}
            variables = envelope.get("variables") or {}

            recipient_email = user.get("email")
            if not recipient_email:
                self.status_store.update_status(request_id, "failed", settings.provider_name, "missing email")
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
                return

            slug = template_info.get("slug")
            locale = user.get("locale") or template_info.get("locale")
            template = self.template_client.get_active_template(slug, locale)
            rendered_subject = self._render(template["subject"], variables)
            rendered_body = self._render(template["body"], variables)

            success = asyncio.run(
                self.email_sender.send_raw_email(recipient_email, rendered_subject, rendered_body)
            )

            if success:
                self.status_store.update_status(request_id, "delivered", settings.provider_name, None)
                ch.basic_ack(delivery_tag=method.delivery_tag)
            else:
                self.status_store.update_status(request_id, "failed", settings.provider_name, "smtp failure")
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
        except Exception as exc:
            logger.error("Email processing failed: %s", exc)
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
