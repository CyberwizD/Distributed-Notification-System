# app/email_sender.py
import jinja2
import aiosmtplib
import os
from email.message import EmailMessage
from dotenv import load_dotenv, find_dotenv
from app.config.settings import settings
import ssl
import asyncio
from typing import Optional, Dict, Any
import logging
from pathlib import Path

logger = logging.getLogger(__name__)
load_dotenv(find_dotenv())

class EmailSender:
    def __init__(self):
        # use settings values (with sensible defaults)
        self.smtp_host: str = settings.smtp_host or "smtp.gmail.com"
        self.smtp_port: int = int(settings.smtp_port or 465)
        self.smtp_username: Optional[str] = settings.smtp_username
        self.smtp_password: Optional[str] = settings.smtp_password

        # Initialize templates (keep builtin templates if you have them)
        self.template_env = jinja2.Environment(
            loader=jinja2.DictLoader(self._get_builtin_templates() if hasattr(self, "_get_builtin_templates") else {})
        )

    # new: render template by id using Jinja2 (txt/html)
    def render_template(self, template_id: str, variables: Optional[Dict[str, Any]] = None) -> (str, Optional[str]):
        variables = variables or {}
        text_out = None
        html_out = None
        # try templates loaded in the Jinja environment first
        for ext in ("txt", "html"):
            name = f"{template_id}.{ext}"
            try:
                tpl = self.template_env.get_template(name)
                rendered = tpl.render(**variables)
                if ext == "txt":
                    text_out = rendered
                else:
                    html_out = rendered
            except jinja2.TemplateNotFound:
                continue
        # fallback: look for templates on disk ...
        if not text_out or not html_out:
            tpl_dir = Path(__file__).resolve().parents[1] / "templates"
            if tpl_dir.exists():
                txt_path = tpl_dir / f"{template_id}.txt"
                html_path = tpl_dir / f"{template_id}.html"
                if txt_path.exists() and not text_out:
                    text_out = jinja2.Template(txt_path.read_text()).render(**variables)
                if html_path.exists() and not html_out:
                    html_out = jinja2.Template(html_path.read_text()).render(**variables)
        # debug log so you can see whether rendering produced content
        logger.debug("render_template %s -> text=%d html=%d", template_id, len(text_out or ""), len(html_out or ""))
        return (text_out or "", html_out)

    async def send_message(
        self,
        subject: str,
        recipient: str,
        body_text: str,
        body_html: Optional[str] = None,
        extra_headers: Optional[Dict[str, Any]] = None,
    ) -> None:
        msg = EmailMessage()
        msg["From"] = self.smtp_username or "no-reply@example.com"
        msg["To"] = recipient
        msg["Subject"] = subject
        if extra_headers:
            for k, v in extra_headers.items():
                msg[k] = v

        # set content
        if body_text:
            msg.set_content(body_text)
        if body_html:
            msg.add_alternative(body_html, subtype="html")
        # Guard: don't send completely empty email (use get_body for html)
        plain = (body_text or "").strip()
        html_alt = ""
        html_part = msg.get_body(preferencelist=('html',))
        if html_part is not None:
            try:
                html_alt = html_part.get_content().strip()
            except Exception:
                html_alt = ""
        if not plain and not html_alt:
            logger.error("Refusing to send empty email to %s (subject=%s)", recipient, subject)
            raise ValueError("Empty email body")

        # Check if SMTP is configured
        if not self.smtp_username or not self.smtp_password:
            logger.warning("SMTP not configured - email would be sent to: %s", recipient)
            logger.info("SUBJECT: %s", subject)
            logger.info("BODY TEXT: %s", body_text)
            logger.info("BODY HTML: %s", body_html)
            return

        use_ssl = self.smtp_port == 465         # implicit TLS (SMTPS)
        use_starttls = self.smtp_port == 587    # STARTTLS

        # create SMTP client
        smtp = aiosmtplib.SMTP(hostname=self.smtp_host, port=self.smtp_port, use_tls=use_ssl, timeout=30)

        await smtp.connect()
        if use_starttls:
            # start TLS if required
            await smtp.starttls()
        if self.smtp_username and self.smtp_password:
            await smtp.login(self.smtp_username, self.smtp_password)
        await smtp.send_message(msg)
        await smtp.quit()

    # Add this adapter so existing callers using send_email keep working
    async def send_email(
        self,
        *args,
        **kwargs,
    ) -> None:
        """
        Backwards-compatible adapter.
        Supports: recipient_email / to / to_email / recipient_email,
        Accepts template_id and variables to render body before sending.
        """
        # legacy recipient names
        recipient = kwargs.pop("to_email", None) or kwargs.pop("to", None) or kwargs.pop("recipient", None) or kwargs.pop("recipient_email", None)
        template_id = kwargs.pop("template_id", None)
        variables = kwargs.pop("variables", None) or {}

        # subject/body extraction (as before)
        subject = kwargs.pop("subject", None)
        body_text = kwargs.pop("body_text", None) or kwargs.pop("body", None)
        body_html = kwargs.pop("body_html", None) or kwargs.pop("html", None)
        extra_headers = kwargs.pop("extra_headers", None) or kwargs.pop("headers", None)

        # positional args heuristic
        if not recipient:
            if len(args) >= 1 and "@" in str(args[0]):
                recipient = args[0]
        if not subject:
            if len(args) >= 1 and recipient and args[0] != recipient:
                subject = args[0]
            elif len(args) >= 2 and recipient:
                subject = args[1]
        if not body_text:
            if len(args) >= 2 and recipient and args[1] != subject:
                body_text = args[1]
            elif len(args) >= 3:
                body_text = args[2]

        if template_id:
            try:
                rendered_text, rendered_html = self.render_template(template_id, variables)
                # prefer template outputs if present
                if rendered_text and not body_text:
                    body_text = rendered_text
                if rendered_html and not body_html:
                    body_html = rendered_html
            except Exception as e:
                logger.exception("Template render error for %s: %s", template_id, e)
                # continue — send fallback body if any

        if not recipient:
            raise ValueError("recipient email not provided (to_email / to / recipient / recipient_email)")

        subject = subject or "No subject"
        body_text = body_text or ""

        return await self.send_message(
            subject=subject,
            recipient=recipient,
            body_text=body_text,
            body_html=body_html,
            extra_headers=extra_headers,
        )

    # synchronous helper to call from sync code
    def send_message_sync(self, *args, **kwargs):
        return asyncio.run(self.send_message(*args, **kwargs))