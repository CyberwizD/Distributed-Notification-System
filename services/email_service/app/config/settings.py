import os
from pathlib import Path
from dotenv import load_dotenv, find_dotenv
from pydantic_settings import BaseSettings
from typing import Optional
from pydantic import field_validator

# load .env for local dev only (no-op if not present)
load_dotenv(find_dotenv())

def read_file_if_env(file_env: str) -> Optional[str]:
    path = os.getenv(file_env)
    if not path:
        return None
    try:
        # If the env points directly to a file (e.g. /run/secrets/...) read it
        return Path(path).read_text().strip()
    except Exception:
        # If path isn't readable, ignore and fallback to normal env
        return None

class Settings(BaseSettings):
    # RabbitMQ Settings
    rabbitmq_url: str = os.getenv("RABBITMQ_URL", "amqp://admin:admin123@rabbitmq:5672/")
    email_queue: str = os.getenv("EMAIL_QUEUE", "email_queue")
    failed_queue: str = "failed.queue"

    # Service
    service_name: str = os.getenv("SERVICE_NAME", "email-service")
    service_port: int = int(os.getenv("SERVICE_PORT", "2525"))

    # SMTP settings (can be provided directly or via *_FILE pointing to secret file)
    smtp_host: Optional[str] = os.getenv("SMTP_HOST", "smtp.gmail.com")
    smtp_port: Optional[int] = None
    smtp_username: Optional[str] = None
    smtp_password: Optional[str] = None

    class Config:
        env_file = ".env"
        case_sensitive = False

    @field_validator("smtp_port", mode="before")
    def _smtp_port_from_env(cls, v):
        # prefer explicit env SMTP_PORT, else read from env var string
        val = os.getenv("SMTP_PORT")
        if val:
            try:
                return int(val)
            except Exception:
                return v
        return v

    @field_validator("smtp_username", mode="before")
    def _smtp_username_from_file(cls, v):
        file_val = read_file_if_env("SMTP_USERNAME_FILE")
        if file_val:
            return file_val
        return os.getenv("SMTP_USERNAME") or v

    @field_validator("smtp_password", mode="before")
    def _smtp_password_from_file(cls, v):
        file_val = read_file_if_env("SMTP_PASSWORD_FILE")
        if file_val:
            return file_val
        return os.getenv("SMTP_PASSWORD") or v

settings = Settings()