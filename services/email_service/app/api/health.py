from fastapi import APIRouter
from datetime import datetime
from app.config.settings import settings

router = APIRouter()

@router.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": settings.service_name,
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0"
    }