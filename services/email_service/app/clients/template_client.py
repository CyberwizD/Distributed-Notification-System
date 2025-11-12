import httpx
from typing import Optional, Dict, Any
import logging

from app.config.settings import settings

logger = logging.getLogger(__name__)


class TemplateClient:
    def __init__(self) -> None:
        self.base_url = settings.template_service_url.rstrip("/")

    def get_active_template(self, slug: str, locale: Optional[str]) -> Dict[str, Any]:
        if not slug:
            raise ValueError("template slug/code is required")

        params = {}
        if locale:
            params["locale"] = locale

        url = f"{self.base_url}/v1/templates/{slug}/active"
        logger.debug("Fetching template from %s", url)

        with httpx.Client(timeout=5.0) as client:
            response = client.get(url, params=params)
            response.raise_for_status()

        payload = response.json()
        if not payload.get("success"):
            raise RuntimeError(payload.get("message") or "template service error")

        data = payload.get("data") or {}
        return {
            "subject": data.get("subject") or slug,
            "body": data.get("body") or "",
        }
