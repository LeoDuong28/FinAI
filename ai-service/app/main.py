"""FastAPI application entry point."""

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.config import settings
from app.models.schemas import HealthResponse
from app.routers import categorize, chat, insights, negotiate, subscriptions
from app.services.categorizer import is_finbert_available
from app.utils.gemini_client import gemini_client

logging.basicConfig(
    level=logging.DEBUG if settings.debug else logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan: startup and shutdown events."""
    logger.info("AI service starting up")
    logger.info("Gemini API available: %s", gemini_client.is_available)
    # FinBERT is lazy-loaded on first request to speed up startup
    yield
    logger.info("AI service shutting down")


app = FastAPI(
    title="FinAI - AI Service",
    version="1.0.0",
    docs_url="/docs" if settings.debug else None,
    redoc_url=None,
    lifespan=lifespan,
)

# Register routers
app.include_router(categorize.router, prefix="/api", tags=["categorize"])
app.include_router(
    subscriptions.router, prefix="/api", tags=["subscriptions"]
)
app.include_router(insights.router, prefix="/api", tags=["insights"])
app.include_router(negotiate.router, prefix="/api", tags=["negotiate"])
app.include_router(chat.router, prefix="/api", tags=["chat"])


@app.get("/health", response_model=HealthResponse)
async def health() -> HealthResponse:
    """Health check endpoint."""
    return HealthResponse(
        status="healthy",
        version="1.0.0",
        services={
            "finbert": "up" if is_finbert_available() else "not_loaded",
            "gemini": "up" if gemini_client.is_available else "not_configured",
        },
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "app.main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
    )
