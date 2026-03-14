from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Service
    host: str = "0.0.0.0"
    port: int = 8081
    debug: bool = False

    # Authentication
    service_api_key: str = ""
    service_hmac_secret: str = ""

    # Google Gemini
    gemini_api_key: str = ""
    gemini_model: str = "gemini-2.0-flash"
    gemini_max_rpm: int = 15

    # FinBERT
    finbert_model: str = "ProsusAI/finbert"
    finbert_confidence_threshold: float = 0.85

    # Cache
    cache_ttl_categorization: int = 86400  # 24 hours
    cache_ttl_negotiation: int = 604800  # 7 days
    cache_ttl_forecast: int = 3600  # 1 hour
    cache_max_entries: int = 10000

    model_config = {"env_prefix": "AI_"}


settings = Settings()
