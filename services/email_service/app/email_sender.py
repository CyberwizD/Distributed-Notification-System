# app/email_sender.py
import aiosmtplib
import jinja2
import os
from typing import Dict, Any
import logging
import asyncio
from email.message import EmailMessage
from dotenv import load_dotenv, find_dotenv

logger = logging.getLogger(__name__)

class EmailSender:
    def __init__(self):
        load_dotenv(find_dotenv())
        # Gmail SMTP settings
        self.smtp_host = os.getenv("SMTP_HOST", "smtp.gmail.com")
        self.smtp_port = int(os.getenv("SMTP_PORT") or "465")
        self.smtp_username = os.getenv("SMTP_USERNAME")
        self.smtp_password = os.getenv("SMTP_PASSWORD")
        
        # Initialize templates
        self.template_env = jinja2.Environment(
            loader=jinja2.DictLoader(self._get_builtin_templates())
        )
    
    def _get_builtin_templates(self) -> Dict[str, str]:
        """Built-in email templates"""
        return {
            "welcome": """
            <!DOCTYPE html>
            <html>
            <head>
                <style>
                    body { font-family: Arial, sans-serif; margin: 20px; }
                    .container { max-width: 600px; margin: 0 auto; }
                    .header { background: #4F46E5; color: white; padding: 20px; text-align: center; }
                    .content { padding: 20px; background: #f9f9f9; }
                    .footer { padding: 20px; text-align: center; color: #666; }
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="header">
                        <h1>Welcome to Our Service! 🎉</h1>
                    </div>
                    <div class="content">
                        <h2>Hello {{name}},</h2>
                        <p>Thank you for joining our notification service. We're excited to have you on board!</p>
                        <p><strong>Your verification code: {{verification_code}}</strong></p>
                        <p>Use this code to verify your account and get started.</p>
                    </div>
                    <div class="footer">
                        <p>Best regards,<br>The Notification Team</p>
                    </div>
                </div>
            </body>
            </html>
            """
        }
    
    def render_template(self, template_id: str, variables: Dict[str, Any]) -> str:
        """Render email template with variables"""
        try:
            template = self.template_env.get_template(template_id)
            return template.render(**variables)
        except jinja2.TemplateError as e:
            logger.warning(f"Template {template_id} not found, using fallback: {e}")
            return f"""
            <html>
            <body>
                <h2>Notification</h2>
                <p>{variables.get('message', 'You have a new notification from our service.')}</p>
            </body>
            </html>
            """
    
    async def send_email(self, to_email: str, subject: str, template_id: str, variables: Dict[str, Any]) -> bool:
        """Send email using SMTP - try SSL first, fallback to STARTTLS on port 587"""
        try:
            # Check if SMTP is configured
            if not self.smtp_username or not self.smtp_password:
                logger.warning("SMTP not configured - email would be sent to: %s", to_email)
                logger.info("SUBJECT: %s", subject)
                logger.info("TEMPLATE: %s", template_id)
                logger.info("VARIABLES: %s", variables)
                return True

            logger.info("🔧 Attempting to send email via SMTP...")
            logger.info(f"🔧 SMTP Config: {self.smtp_host}:{self.smtp_port}")
            logger.info(f"🔧 Username: {self.smtp_username}")

            # Create message
            message = EmailMessage()
            message["From"] = self.smtp_username
            message["To"] = to_email
            message["Subject"] = subject

            # Render HTML content
            html_content = self.render_template(template_id, variables)
            message.set_content(html_content, subtype='html')

            # Helper to cleanly quit smtp if created
            smtp = None
            try:
                # Try SSL/TLS first (typical for port 465)
                logger.info("🔧 Connecting using SSL/TLS...")
                smtp = aiosmtplib.SMTP(hostname=self.smtp_host, port=self.smtp_port, use_tls=True)
                await asyncio.wait_for(smtp.connect(), timeout=10)
                logger.info("🔧 Connected (SSL), attempting login...")
                await asyncio.wait_for(smtp.login(self.smtp_username, self.smtp_password), timeout=10)
                logger.info("🔧 Login successful, sending message...")
                await asyncio.wait_for(smtp.send_message(message), timeout=20)
                logger.info("✅ Email sent successfully (SSL) to: %s", to_email)
                return True
            except Exception as ssl_err:
                logger.warning("⚠️ SSL/TLS send failed: %s", ssl_err)
                # Ensure any partial connection is closed
                try:
                    if smtp is not None:
                        await smtp.quit()
                except Exception:
                    pass

                # Try STARTTLS on port 587 as a fallback
                try:
                    logger.info("🔧 Attempting STARTTLS fallback on port 587...")
                    smtp = aiosmtplib.SMTP(hostname=self.smtp_host, port=587, use_tls=False)
                    await asyncio.wait_for(smtp.connect(), timeout=10)
                    logger.info("🔧 Connected (no TLS), starting STARTTLS...")
                    await asyncio.wait_for(smtp.starttls(), timeout=10)
                    logger.info("🔧 STARTTLS established, attempting login...")
                    await asyncio.wait_for(smtp.login(self.smtp_username, self.smtp_password), timeout=10)
                    logger.info("🔧 Login successful, sending message (STARTTLS)...")
                    await asyncio.wait_for(smtp.send_message(message), timeout=20)
                    logger.info("✅ Email sent successfully (STARTTLS) to: %s", to_email)
                    return True
                except Exception as starttls_err:
                    logger.error("❌ STARTTLS send failed: %s", starttls_err)
                    try:
                        if smtp is not None:
                            await smtp.quit()
                    except Exception:
                        pass
                    return False
            finally:
                try:
                    if smtp is not None:
                        await smtp.quit()
                except Exception:
                    pass

        except asyncio.TimeoutError:
            logger.error("❌ SMTP operation timed out")
            return False
        except Exception as e:
            logger.error(f"❌ SMTP Error: {str(e)}")
            logger.error(f"❌ Error type: {type(e).__name__}")
            return False