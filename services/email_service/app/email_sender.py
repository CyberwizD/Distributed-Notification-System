# app/email_sender.py
import logging
from email.message import EmailMessage
from typing import Any, Dict

import aiosmtplib
import jinja2

from app.config.settings import settings

logger = logging.getLogger(__name__)


class EmailSender:
    def __init__(self) -> None:
        self.smtp_host = settings.smtp_host
        self.smtp_port = settings.smtp_port
        self.smtp_username = settings.smtp_username
        self.smtp_password = settings.smtp_password

        self.template_env = jinja2.Environment(
            loader=jinja2.DictLoader(self._get_builtin_templates())
        )

    def _get_builtin_templates(self) -> Dict[str, str]:
        return {
            "welcome": """
            <!DOCTYPE html>
            <html>
            <body>
                <h2>Welcome to our service, {{name}}!</h2>
                <p>Your verification code is <strong>{{verification_code}}</strong>.</p>
            </body>
            </html>
            """
        }

    def render_template(self, template_id: str, variables: Dict[str, Any]) -> str:
        try:
            template = self.template_env.get_template(template_id)
            return template.render(**variables)
        except jinja2.TemplateError as exc:
            logger.warning("Template %s not found (%s); using fallback", template_id, exc)
            return f"""
            <html>
                <body>
                    <p>{variables.get('message', 'You have a new notification.')}</p>
                </body>
            </html>
            """

    def _is_smtp_configured(self) -> bool:
        return all([self.smtp_host, self.smtp_port, self.smtp_username, self.smtp_password])

    async def send_raw_email(self, to_email: str, subject: str, body: str) -> bool:
        if not self._is_smtp_configured():
            logger.warning("SMTP not configured - simulated send to %s", to_email)
            logger.info("SUBJECT: %s", subject)
            logger.info("BODY: %s", body)
            return True

        message = EmailMessage()
        message["From"] = self.smtp_username
        message["To"] = to_email
        message["Subject"] = subject
        message.set_content(body, subtype="html")

        try:
            await aiosmtplib.send(
                message,
                hostname=self.smtp_host,
                port=self.smtp_port,
                username=self.smtp_username,
                password=self.smtp_password,
                start_tls=True,
            )
            logger.info("Email sent successfully to %s", to_email)
            return True
        except Exception as exc:
            logger.error("Failed to send email via SMTP: %s", exc)
            return False

    async def send_email(
        self, to_email: str, subject: str, template_id: str, variables: Dict[str, Any]
    ) -> bool:
        rendered_body = self.render_template(template_id, variables)
        final_subject = subject or "Notification"
        return await self.send_raw_email(to_email, final_subject, rendered_body)
