import json, time, pika, logging
from app.config.settings import settings
from app.email_sender import EmailSender

class BaseConsumer:
    def __init__(self):
        self.logger = logging.getLogger(__name__)
        self.email_sender = EmailSender()
        self.connection = None
        self.channel = None
    

    def connect(self):
        """Establish connection to RabbitMQ with DLQ settings"""
        try:
            self.connection = pika.BlockingConnection(
                pika.URLParameters(settings.rabbitmq_url)
            )
            self.channel = self.connection.channel()
            
            # Declare main queue with DLX settings
            self.channel.queue_declare(
                queue=settings.email_queue,
                durable=True,
                arguments={
                    'x-dead-letter-exchange': '',
                    'x-dead-letter-routing-key': settings.failed_queue
                }
            )
            
            # Declare failed queue (DLQ)
            self.channel.queue_declare(
                queue=settings.failed_queue,
                durable=True
            )
            
            self.logger.info("✅ Connected to RabbitMQ with DLQ setup")
            return True
            
        except Exception as e:
            self.logger.error(f"❌ Failed to connect to RabbitMQ: {e}")
            return False
    
    def start_consuming(self):
        """Start consuming messages - to be implemented by subclasses"""
        if not self.connect():
            return False
        
        try:
            self.channel.basic_qos(prefetch_count=1)
            self.channel.basic_consume(
                queue=settings.email_queue,
                on_message_callback=self.process_message
            )
            
            self.logger.info(f"🔄 Starting consumer for {settings.email_queue}")
            self.channel.start_consuming()
            
        except Exception as e:
            self.logger.error(f"❌ Consumer error: {e}")
            return False
    
    def process_message(self, ch, method, properties, body):
        """Process message - to be overridden by subclasses"""
        try:
            message = json.loads(body)
            self.logger.info(f"📨 Received message: {message}")
            
            # Acknowledge message (subclasses will implement actual processing)
            ch.basic_ack(delivery_tag=method.delivery_tag)
            
        except Exception as e:
            self.logger.error(f"❌ Message processing error: {e}")
            # Reject and send to DLQ
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
    
    def stop_consuming(self):
        """Stop consuming messages"""
        if self.connection and not self.connection.is_closed:
            self.connection.close()
            self.logger.info("🛑 Consumer stopped")
            
            


class EmailConsumer(BaseConsumer):
    def __init__(self):
        super().__init__()

    def start_consuming(self):
        max_retries = 10
        delay = 1.0
        for attempt in range(1, max_retries + 1):
            try:
                # reuse BaseConsumer.connect() which sets up the queues / DLQ
                if self.connect():
                    self.logger.info("✅ Connected to RabbitMQ")
                    break
            except Exception as exc:
                self.logger.warning("RabbitMQ connect attempt %d/%d failed: %s", attempt, max_retries, exc)

            if attempt == max_retries:
                self.logger.error("❌ Failed to connect to RabbitMQ after %d attempts", max_retries)
                return False

            time.sleep(delay)
            delay = min(delay * 2, 10)

        try:
            # Start consuming messages
            self.channel.basic_qos(prefetch_count=1)
            self.channel.basic_consume(
                queue=settings.email_queue,
                on_message_callback=self.process_message
            )
            self.logger.info(f"🔄 Starting consumer for {settings.email_queue}")
            self.channel.start_consuming()
        except Exception as e:
            self.logger.error(f"❌ Consumer error: {e}")
            return False
    
    def process_message(self, ch, method, properties, body):
        """Process email messages and send actual emails"""
        try:
            message_data = json.loads(body)
            self.logger.info(f"📧 Processing email message: {message_data}")
            
            # Extract email data (accept direct body_text/body_html or template + variables)
            recipient_email = (
                message_data.get("recipient_email")
                or message_data.get("to")
                or message_data.get("to_email")
            )
            template_id = message_data.get("template_id")
            subject = message_data.get("subject", "Notification from Our Service")
            variables = message_data.get("variables", {}) or {}
            body_text = message_data.get("body_text") or message_data.get("body")
            body_html = message_data.get("body_html") or message_data.get("html")

            if not recipient_email:
                self.logger.error("❌ No recipient email provided")
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
                return

            # Send actual email (do not ack until successful)
            import asyncio
            try:
                # Pass both template info and any direct body; EmailSender should prefer body_text if present
                asyncio.run(
                    self.email_sender.send_email(
                        recipient_email=recipient_email,
                        subject=subject,
                        template_id=template_id,
                        variables=variables,
                        body_text=body_text,
                        body_html=body_html,
                    )
                )
            except Exception as send_err:
                self.logger.exception("❌ Failed to send email to %s: %s", recipient_email, send_err)

                # publish original message + error to failed queue (DLQ)
                try:
                    failure_payload = {
                        "original_message": message_data,
                        "error": str(send_err),
                        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%S")
                    }
                    self.channel.basic_publish(
                        exchange='',
                        routing_key=settings.failed_queue,
                        body=json.dumps(failure_payload),
                        properties=pika.BasicProperties(delivery_mode=2)  # persistent
                    )
                    self.logger.info("➡️ Published failed message to DLQ '%s'", settings.failed_queue)
                except Exception as pub_err:
                    self.logger.exception("❌ Failed to publish to DLQ: %s", pub_err)

                # nack original message (do not requeue)
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
                return

            # success — acknowledge original message
            self.logger.info(f"✅ Email sent successfully to {recipient_email}")
            ch.basic_ack(delivery_tag=method.delivery_tag)

        except Exception as e:
            self.logger.exception("❌ Email processing failed: %s", e)
            try:
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
            except Exception:
                pass