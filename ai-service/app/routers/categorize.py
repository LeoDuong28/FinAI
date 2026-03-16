"""Transaction categorization endpoint."""

from fastapi import APIRouter, Depends

from app.models.schemas import CategorizeRequest, CategorizeResponse
from app.services.categorizer import categorize_transactions
from app.utils.auth import verify_service_auth

router = APIRouter()


@router.post("/categorize", response_model=CategorizeResponse)
async def categorize(
    request: CategorizeRequest,
    _: None = Depends(verify_service_auth),
) -> CategorizeResponse:
    """Categorize a batch of transactions."""
    results = await categorize_transactions(request.transactions)
    return CategorizeResponse(categories=results)
