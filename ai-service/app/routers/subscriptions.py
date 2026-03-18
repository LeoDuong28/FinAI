"""Subscription detection endpoint."""

import asyncio

from fastapi import APIRouter, Depends

from app.models.schemas import (
    DetectSubscriptionsRequest,
    DetectSubscriptionsResponse,
)
from app.services.subscription_detector import detect_subscriptions
from app.utils.auth import verify_service_auth

router = APIRouter()


@router.post(
    "/detect-subscriptions", response_model=DetectSubscriptionsResponse
)
async def detect(
    request: DetectSubscriptionsRequest,
    _: None = Depends(verify_service_auth),
) -> DetectSubscriptionsResponse:
    """Detect recurring subscriptions from transaction history."""
    subs = await asyncio.to_thread(detect_subscriptions, request.transactions)
    return DetectSubscriptionsResponse(subscriptions=subs)
