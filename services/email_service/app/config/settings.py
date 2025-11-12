import os
from pathlib import Path
from dotenv import load_dotenv, find_dotenv
from pydantic_settings import BaseSettings
from typing import Optional

# load .env for local dev only (no-op if not present)
load_dotenv(find_dotenv())

def read_secret_env(var_name: str, file_var_name: str):
    file_path = os.getenv(file_var_name)
    if file_path:
        # file_path may be an absolute path (e.g. /run/secrets/smtp_username)
        try:
            return Path(file_path).read_text().strip()
        except Exception:
            # fallback: try reading by secret name from /run/secrets/<name>
            try:
                secret_name = Path(file_path).name
                return Path(f"/run/secrets/{secret_name}").read_text().strip()
            except Exception:
                return None
    return os.getenv(var_name)

class Settings(BaseSettings):
    # RabbitMQ Settings
    rabbitmq_url: str = os.getenv("RABBITMQ_URL", "amqp://admin:admin123@rabbitmq:5672/")
    email_queue: str = os.getenv("EMAIL_QUEUE", "email_queue")
    failed_queue: str = "failed.queue"
    
    # Service Settings
    service_name: str = os.getenv("SERVICE_NAME", "email-service")
    service_port: int = 2525
    
    # SMTP Settings (for future use)
    smtp_host: Optional[str] = os.getenv("SMTP_HOST", "smtp.gmail.com")
    smtp_port: Optional[int] = int(os.getenv("SMTP_PORT") or "465")
    smtp_username: Optional[str] = read_secret_env("SMTP_USERNAME", "SMTP_USERNAME_FILE")
    smtp_password: Optional[str] = read_secret_env("SMTP_PASSWORD", "SMTP_PASSWORD_FILE")
    
    class Config:
        env_file = ".env"

settings = Settings()