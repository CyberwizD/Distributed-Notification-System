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

        # Initialize templates with built-in templates
        self.template_env = jinja2.Environment(
            loader=jinja2.DictLoader(self._get_builtin_templates())
        )

    def _get_builtin_templates(self):
        """Provide built-in email templates to prevent empty email bodies"""
        return {
            "welcome.txt": """Hello {{ name }},

Your verification code is: {{ verification_code }}

This code will expire in 15 minutes.

If you didn't request this, please ignore this email.

Thank you!
The Team""",

            "welcome.html": """<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .code { font-size: 24px; font-weight: bold; color: #2563eb; padding: 10px; background: #f3f4f6; border-radius: 5px; text-align: center; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; color: #6b7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Hello {{ name }}!</h2>
        <p>Your verification code is:</p>
        <div class="code">{{ verification_code }}</div>
        <p>This code will expire in 15 minutes.</p>
        <p>If you didn't request this verification, please ignore this email.</p>
        <div class="footer">
            <p>Thank you!<br>The Team</p>
        </div>
    </div>
</body>
</html>""",

            "notification.txt": """Hello {{ name }},

{{ message }}

Thank you!
The Team""",

            "notification.html": """<!DOCTYPE html>
<html>
<body>
    <h2>Hello {{ name }}!</h2>
    <p>{{ message }}</p>
    <p>Thank you!<br>The Team</p>
</body>
</html>""",

            "default.txt": """Hello {{ name }},

This is a notification from our service.

{% if verification_code %}
Your verification code: {{ verification_code }}
{% endif %}

Thank you!
The Team""",

            "default.html": """<!DOCTYPE html>
<html>
<body>
    <h2>Hello {{ name }}!</h2>
    <p>This is a notification from our service.</p>
    {% if verification_code %}
    <p>Your verification code: <strong>{{ verification_code }}</strong></p>
    {% endif %}
    <p>Thank you!<br>The Team</p>
</body>
</html>"""
        }

    def render_template(self, template_id: str, variables: Optional[Dict[str, Any]] = None) -> (str, Optional[str]):
        """Render template by id using Jinja2 (txt/html) with fallback logic"""
        variables = variables or {}
        text_out = None
        html_out = None
        
        logger.info(f"🔍 Looking for template: {template_id}")
        
        # Try templates loaded in the Jinja environment first
        for ext in ("txt", "html"):
            name = f"{template_id}.{ext}"
            try:
                tpl = self.template_env.get_template(name)
                rendered = tpl.render(**variables)
                if ext == "txt":
                    text_out = rendered
                else:
                    html_out = rendered
                logger.info(f"✅ Found built-in template: {name}")
            except jinja2.TemplateNotFound:
                logger.debug(f"📭 Built-in template not found: {name}")
                continue
        
        # Fallback: look for templates on disk
        if not text_out or not html_out:
            tpl_dir = Path(__file__).resolve().parents[1] / "templates"
            logger.info(f"📁 Checking template directory: {tpl_dir}")
            
            if tpl_dir.exists():
                for ext in ("txt", "html"):
                    name = f"{template_id}.{ext}"
                    file_path = tpl_dir / name
                    
                    if file_path.exists():
                        logger.info(f"✅ Found file template: {file_path}")
                        template_content = file_path.read_text()
                        tpl = jinja2.Template(template_content)
                        rendered = tpl.render(**variables)
                        if ext == "txt":
                            text_out = rendered
                        else:
                            html_out = rendered
        
        # Final fallback: use default template if requested template not found
        if not text_out and template_id not in ["welcome", "notification", "default"]:
            logger.warning(f"📭 Template '{template_id}' not found, using default template")
            text_out, html_out = self.render_template("default", variables)
        
        # Ensure we have at least text content
        if not text_out:
            logger.error(f"❌ No template content found for {template_id}")
            text_out = f"Hello {variables.get('name', 'User')},\n\n"
            if variables.get('verification_code'):
                text_out += f"Your verification code: {variables['verification_code']}\n\n"
            text_out += "Thank you!\nThe Team"
        
        if not html_out:
            html_out = f"""
            <!DOCTYPE html>
            <html>
            <body>
                <h2>Hello {variables.get('name', 'User')}!</h2>
                {"<p>Your verification code: <strong>" + variables['verification_code'] + "</strong></p>" if variables.get('verification_code') else ""}
                <p>Thank you!<br>The Team</p>
            </body>
            </html>
            """
        
        logger.info(f"📧 Template rendering complete - text: {len(text_out)} chars, html: {len(html_out)} chars")
        return (text_out, html_out)

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

        # Set content - ensure we have at least plain text
        body_text = body_text or "No content provided"
        msg.set_content(body_text)
        
        if body_html:
            msg.add_alternative(body_html, subtype="html")
        else:
            # Create simple HTML fallback from text
            simple_html = f"<html><body><pre>{body_text}</pre></body></html>"
            msg.add_alternative(simple_html, subtype="html")

        # Validate we have content before sending
        plain_content = (body_text or "").strip()
        if not plain_content:
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

        # Create SMTP client
        smtp = aiosmtplib.SMTP(hostname=self.smtp_host, port=self.smtp_port, use_tls=use_ssl, timeout=30)

        try:
            await smtp.connect()
            if use_starttls:
                await smtp.starttls()
            
            if self.smtp_username and self.smtp_password:
                await smtp.login(self.smtp_username, self.smtp_password)
            
            await smtp.send_message(msg)
            logger.info(f"✅ Email sent successfully to {recipient}")
            
        except Exception as e:
            logger.error(f"❌ Failed to send email to {recipient}: {e}")
            raise
        finally:
            await smtp.quit()

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
        # Legacy recipient names
        recipient = (kwargs.pop("to_email", None) or 
                    kwargs.pop("to", None) or 
                    kwargs.pop("recipient", None) or 
                    kwargs.pop("recipient_email", None))
        
        template_id = kwargs.pop("template_id", None)
        variables = kwargs.pop("variables", None) or {}

        # Subject/body extraction
        subject = kwargs.pop("subject", None)
        body_text = kwargs.pop("body_text", None) or kwargs.pop("body", None)
        body_html = kwargs.pop("body_html", None) or kwargs.pop("html", None)
        extra_headers = kwargs.pop("extra_headers", None) or kwargs.pop("headers", None)

        # Positional args heuristic
        if not recipient and len(args) >= 1 and "@" in str(args[0]):
            recipient = args[0]
        if not subject and len(args) >= 1 and recipient and args[0] != recipient:
            subject = args[0]
        elif not subject and len(args) >= 2 and recipient:
            subject = args[1]
        if not body_text and len(args) >= 2 and recipient and args[1] != subject:
            body_text = args[1]
        elif not body_text and len(args) >= 3:
            body_text = args[2]

        # If template_id provided, render templates
        if template_id:
            try:
                rendered_text, rendered_html = self.render_template(template_id, variables)
                # Use template outputs if not explicitly provided
                if not body_text:
                    body_text = rendered_text
                if not body_html:
                    body_html = rendered_html
                logger.info(f"✅ Template '{template_id}' rendered successfully")
            except Exception as e:
                logger.error("Template render error for %s: %s", template_id, e)
                # Don't fail completely - use fallback content

        # Validate required fields
        if not recipient:
            raise ValueError("recipient email not provided (to_email / to / recipient / recipient_email)")

        subject = subject or "Notification from Our Service"
        
        # Ensure we have at least some body content
        if not body_text:
            body_text = f"Hello {variables.get('name', 'User')},\n\n"
            if variables.get('verification_code'):
                body_text += f"Your verification code: {variables['verification_code']}\n\n"
            body_text += "Thank you!\nThe Team"
            logger.warning("📝 Using fallback body content")

        return await self.send_message(
            subject=subject,
            recipient=recipient,
            body_text=body_text,
            body_html=body_html,
            extra_headers=extra_headers,
        )

    # Synchronous helper to call from sync code
    def send_message_sync(self, *args, **kwargs):
        return asyncio.run(self.send_message(*args, **kwargs))