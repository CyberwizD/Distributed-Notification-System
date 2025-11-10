from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # RabbitMQ Settings
    rabbitmq_url: str = "amqp://guest:guest@rabbitmq:5672/"
    email_queue: str = "email.queue"
    failed_queue: str = "failed.queue"
    
    # Service Settings
    service_name: str = "email-service"
    service_port: int = 2525
    
    # SMTP Settings (for future use)
    smtp_host: Optional[str] = None
    smtp_port: Optional[int] = 587
    smtp_username: Optional[str] = None
    smtp_password: Optional[str] = None
    
    class Config:
        env_file = ".env"

settings = Settings()