"""Request authentication for inter-service communication."""

import hashlib
import hmac
import logging
import time

from fastapi import Header, HTTPException, Request

from app.config import settings

logger = logging.getLogger(__name__)

# Maximum allowed clock skew for request timestamps (seconds)
_MAX_TIMESTAMP_SKEW = 60  # 1 minute


async def verify_service_auth(
    request: Request,
    x_service_key: str = Header(default=""),
    x_signature: str = Header(default=""),
    x_timestamp: str = Header(default=""),
) -> None:
    """Verify API key and HMAC signature on requests from Go backend."""
    # Fail closed: if no API key is configured, reject all requests
    if not settings.service_api_key:
        logger.error("AI_SERVICE_API_KEY not configured — rejecting request")
        raise HTTPException(status_code=401, detail="Unauthorized")

    if not hmac.compare_digest(x_service_key, settings.service_api_key):
        logger.warning(
            "Invalid service API key from %s",
            request.client.host if request.client else "unknown",
        )
        raise HTTPException(status_code=401, detail="Unauthorized")

    # HMAC signature verification (required when secret is configured)
    if settings.service_hmac_secret:
        if not x_signature or not x_timestamp:
            raise HTTPException(status_code=401, detail="Unauthorized")

        # Replay protection: reject requests with stale timestamps
        try:
            req_time = int(x_timestamp)
        except ValueError:
            raise HTTPException(status_code=401, detail="Unauthorized")

        now = int(time.time())
        if abs(now - req_time) > _MAX_TIMESTAMP_SKEW:
            logger.warning("Request timestamp too far from server time")
            raise HTTPException(status_code=401, detail="Unauthorized")

        body = await request.body()
        # Sign: timestamp + "." + body (matches Go client)
        payload = x_timestamp.encode() + b"." + body
        expected = hmac.new(
            settings.service_hmac_secret.encode(),
            payload,
            hashlib.sha256,
        ).hexdigest()
        if not hmac.compare_digest(x_signature, expected):
            logger.warning("Invalid HMAC signature")
            raise HTTPException(status_code=401, detail="Unauthorized")
