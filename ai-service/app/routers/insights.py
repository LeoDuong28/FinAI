"""Insights endpoints: anomaly detection and spending forecast."""

import asyncio

from fastapi import APIRouter, Depends

from app.models.schemas import (
    AnomalyRequest,
    AnomalyResponse,
    ForecastRequest,
    ForecastResponse,
)
from app.services.anomaly_detector import detect_anomalies
from app.services.forecaster import forecast_spending
from app.utils.auth import verify_service_auth

router = APIRouter()


@router.post("/anomaly-detect", response_model=AnomalyResponse)
async def anomaly_detect(
    request: AnomalyRequest,
    _: None = Depends(verify_service_auth),
) -> AnomalyResponse:
    """Detect anomalous transactions."""
    anomalies = await asyncio.to_thread(
        detect_anomalies, request.transactions, request.history
    )
    return AnomalyResponse(anomalies=anomalies)


@router.post("/forecast", response_model=ForecastResponse)
async def forecast(
    request: ForecastRequest,
    _: None = Depends(verify_service_auth),
) -> ForecastResponse:
    """Forecast spending for the next N days."""
    result = await asyncio.to_thread(
        forecast_spending, request.history, request.days
    )
    return ForecastResponse(forecast=result)
