import pika
import json
import logging
from app.config.settings import settings

class BaseConsumer:
    def __init__(self):
        self.connection = None
        self.channel = None
        self.logger = logging.getLogger(__name__)
    
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