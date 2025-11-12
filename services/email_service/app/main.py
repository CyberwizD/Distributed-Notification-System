from fastapi import FastAPI, HTTPException, BackgroundTasks
from pydantic import BaseModel, EmailStr
from typing import Dict, Any, Optional, List
from datetime import datetime
import threading
import logging
import json
import pika
from app.api.health import router as health_router
from app.consumers.base_consumer import EmailConsumer
from app.config.settings import settings

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title=settings.service_name)

# Request/Response Models
class EmailRequest(BaseModel):
    recipient_email: EmailStr
    template_id: str
    subject: str
    variables: Dict[str, Any]
    idempotency_key: Optional[str] = None

class EmailResponse(BaseModel):
    success: bool
    message_id: str
    message: str
    error: Optional[str] = None
    timestamp: str

class BatchEmailRequest(BaseModel):
    emails: List[EmailRequest]
    idempotency_key: Optional[str] = None

class BatchEmailResponse(BaseModel):
    success: bool
    processed_count: int
    failed_count: int
    results: List[EmailResponse]
    message: str

# Include routers
app.include_router(health_router)

def publish_to_rabbitmq(message: dict):
    """Publish message to RabbitMQ queue"""
    try:
        connection = pika.BlockingConnection(
            pika.URLParameters(settings.rabbitmq_url)
        )
        channel = connection.channel()
        
        # Ensure queue exists
        channel.queue_declare(
            queue=settings.email_queue,
            durable=True,
            arguments={
                'x-dead-letter-exchange': '',
                'x-dead-letter-routing-key': settings.failed_queue
            }
        )
        
        channel.basic_publish(
            exchange='',
            routing_key=settings.email_queue,
            body=json.dumps(message),
            properties=pika.BasicProperties(
                delivery_mode=2,  # make message persistent
            )
        )
        
        connection.close()
        return True
        
    except Exception as e:
        logger.error(f"❌ Failed to publish to RabbitMQ: {e}")
        return False

# Client Endpoints
@app.post("/send-email", response_model=EmailResponse)
async def send_email(request: EmailRequest, background_tasks: BackgroundTasks):
    """Send a single email - Client facing endpoint"""
    try:
        message_id = f"email-{datetime.utcnow().strftime('%Y%m%d%H%M%S')}-{hash(request.recipient_email) % 10000:04d}"
        
        # Create message for RabbitMQ
        message = {
            "recipient_email": request.recipient_email,
            "template_id": request.template_id,
            "subject": request.subject,
            "variables": request.variables,
            "message_id": message_id
        }
        
        # Publish to RabbitMQ instead of sending directly
        success = publish_to_rabbitmq(message)
        
        if success:
            return EmailResponse(
                success=True,
                message_id=message_id,
                message="Email queued successfully",
                timestamp=datetime.utcnow().isoformat()
            )
        else:
            return EmailResponse(
                success=False,
                message_id=message_id,
                message="Failed to queue email",
                error="RabbitMQ connection failed",
                timestamp=datetime.utcnow().isoformat()
            )
            
    except Exception as e:
        logger.error(f"❌ Error in /send-email: {e}")
        return EmailResponse(
            success=False,
            message_id="error",
            message="Internal server error",
            error=str(e),
            timestamp=datetime.utcnow().isoformat()
        )

@app.post("/send-batch-emails", response_model=BatchEmailResponse)
async def send_batch_emails(request: BatchEmailRequest):
    """Send multiple emails in batch - Client facing endpoint"""
    try:
        results = []
        processed_count = 0
        failed_count = 0
        
        for email_request in request.emails:
            message_id = f"batch-{datetime.utcnow().strftime('%Y%m%d%H%M%S')}-{processed_count:04d}"
            
            try:
                # Create message for RabbitMQ
                message = {
                    "recipient_email": email_request.recipient_email,
                    "template_id": email_request.template_id,
                    "subject": email_request.subject,
                    "variables": email_request.variables,
                    "message_id": message_id
                }
                
                # Publish to RabbitMQ
                success = publish_to_rabbitmq(message)
                
                if success:
                    results.append(EmailResponse(
                        success=True,
                        message_id=message_id,
                        message="Email queued successfully",
                        timestamp=datetime.utcnow().isoformat()
                    ))
                    processed_count += 1
                else:
                    results.append(EmailResponse(
                        success=False,
                        message_id=message_id,
                        message="Failed to queue email",
                        error="RabbitMQ connection failed",
                        timestamp=datetime.utcnow().isoformat()
                    ))
                    failed_count += 1
                    
            except Exception as e:
                results.append(EmailResponse(
                    success=False,
                    message_id=message_id,
                    message="Internal error processing email",
                    error=str(e),
                    timestamp=datetime.utcnow().isoformat()
                ))
                failed_count += 1
        
        return BatchEmailResponse(
            success=failed_count == 0,
            processed_count=processed_count,
            failed_count=failed_count,
            results=results,
            message=f"Processed {processed_count} emails, {failed_count} failed"
        )
        
    except Exception as e:
        logger.error(f"❌ Error in /send-batch-emails: {e}")
        return BatchEmailResponse(
            success=False,
            processed_count=0,
            failed_count=len(request.emails),
            results=[],
            message=f"Batch processing failed: {str(e)}"
        )

@app.post("/test-email")
async def test_email(recipient_email: str = "test@example.com"):
    """Test endpoint with parameter support"""
    try:
        message_id = f"test-{datetime.utcnow().strftime('%Y%m%d%H%M%S')}"
        
        # Create test message for RabbitMQ
        message = {
            "recipient_email": recipient_email,
            "template_id": "welcome",
            "subject": "Test Email from Notification System",
            "variables": {
                "name": "Test User",
                "verification_code": "123456"
            },
            "message_id": message_id
        }
        
        success = publish_to_rabbitmq(message)
        
        return {
            "success": success,
            "message": "Test email queued for processing",
            "sent_to": recipient_email,
            "message_id": message_id
        }
    except Exception as e:
        logger.error(f"❌ Error in /test-email: {e}")
        return {
            "success": False,
            "message": f"Test failed: {str(e)}",
            "sent_to": recipient_email
        }

@app.on_event("startup")
def startup_event():
    """Start the email consumer on startup"""
    try:
        consumer = EmailConsumer()
        thread = threading.Thread(target=consumer.start_consuming, daemon=True)
        thread.start()
        app.state.email_consumer = consumer
        app.state.consumer_thread = thread
        logger.info("🚀 Email consumer thread started")
        
        # Test RabbitMQ connection
        connection = pika.BlockingConnection(pika.URLParameters(settings.rabbitmq_url))
        connection.close()
        logger.info("✅ RabbitMQ connection test successful")
        
    except Exception as e:
        logger.error(f"❌ Startup failed: {e}")

@app.on_event("shutdown")
def shutdown_event():
    """Cleanup on shutdown"""
    consumer = getattr(app.state, "email_consumer", None)
    if consumer:
        try:
            consumer.stop_consuming()
        except Exception:
            pass
    logger.info("🛑 Shutdown complete")

@app.get("/")
async def root():
    return {
        "message": "Email Service Running",
        "service": settings.service_name,
        "endpoints": {
            "health": "/health",
            "send_email": "/send-email (POST)",
            "send_batch_emails": "/send-batch-emails (POST)", 
            "test_email": "/test-email (POST)",
            "docs": "/docs"
        }
    }