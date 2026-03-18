"""Bill negotiation tips endpoint."""

from fastapi import APIRouter, Depends

from app.models.schemas import NegotiateRequest, NegotiateResponse
from app.services.negotiator import get_negotiation_tips
from app.utils.auth import verify_service_auth

router = APIRouter()


@router.post("/negotiate-tips", response_model=NegotiateResponse)
async def negotiate(
    request: NegotiateRequest,
    _: None = Depends(verify_service_auth),
) -> NegotiateResponse:
    """Get AI-generated bill negotiation tips."""
    tips = await get_negotiation_tips(request.bills)
    return NegotiateResponse(tips=tips)
