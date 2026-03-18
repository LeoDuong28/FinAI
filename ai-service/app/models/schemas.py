from datetime import date as _date
from typing import Literal

from pydantic import BaseModel, Field, field_validator


# --- Categorization ---

class TransactionInput(BaseModel):
    id: str
    name: str = Field(max_length=500)
    merchant_name: str | None = Field(default=None, max_length=255)
    amount: float = Field(ge=0)
    type: str = "debit"  # debit | credit


class CategoryResult(BaseModel):
    transaction_id: str
    category: str
    confidence: float = Field(ge=0.0, le=1.0)


class CategorizeRequest(BaseModel):
    transactions: list[TransactionInput] = Field(max_length=50)


class CategorizeResponse(BaseModel):
    categories: list[CategoryResult]


# --- Subscription Detection ---

class TransactionHistory(BaseModel):
    name: str = Field(max_length=500)
    merchant_name: str | None = Field(default=None, max_length=255)
    amount: float = Field(ge=0)
    date: str  # YYYY-MM-DD
    type: str = "debit"

    @field_validator("date")
    @classmethod
    def validate_date(cls, v: str) -> str:
        try:
            _date.fromisoformat(v)
        except ValueError:
            raise ValueError("date must be in YYYY-MM-DD format")
        return v


class DetectedSubscription(BaseModel):
    merchant_name: str
    amount: float
    frequency: str  # weekly | monthly | quarterly | yearly
    confidence: float = Field(ge=0.0, le=1.0)
    next_billing: str | None = None  # YYYY-MM-DD
    last_charged: str | None = None  # YYYY-MM-DD
    transaction_count: int


class DetectSubscriptionsRequest(BaseModel):
    transactions: list[TransactionHistory] = Field(max_length=2000)


class DetectSubscriptionsResponse(BaseModel):
    subscriptions: list[DetectedSubscription]


# --- Anomaly Detection ---

class AnomalyInput(BaseModel):
    id: str
    name: str = Field(max_length=500)
    merchant_name: str | None = Field(default=None, max_length=255)
    amount: float = Field(ge=0)
    date: str  # YYYY-MM-DD
    category: str | None = None

    @field_validator("date")
    @classmethod
    def validate_date(cls, v: str) -> str:
        try:
            _date.fromisoformat(v)
        except ValueError:
            raise ValueError("date must be in YYYY-MM-DD format")
        return v


class AnomalyResult(BaseModel):
    transaction_id: str
    is_anomaly: bool
    reason: str
    z_score: float | None = None


class AnomalyRequest(BaseModel):
    transactions: list[AnomalyInput] = Field(max_length=100)
    history: list[AnomalyInput] = Field(max_length=2000)


class AnomalyResponse(BaseModel):
    anomalies: list[AnomalyResult]


# --- Spending Forecast ---

class DailySpending(BaseModel):
    date: str  # YYYY-MM-DD
    amount: float = Field(ge=0)
    category: str | None = None

    @field_validator("date")
    @classmethod
    def validate_date(cls, v: str) -> str:
        try:
            _date.fromisoformat(v)
        except ValueError:
            raise ValueError("date must be in YYYY-MM-DD format")
        return v


class ForecastResult(BaseModel):
    predicted_total: float
    daily_predictions: list[DailySpending]
    category_breakdown: dict[str, float] | None = None


class ForecastRequest(BaseModel):
    history: list[DailySpending] = Field(max_length=365)
    days: int = Field(default=30, ge=1, le=90)


class ForecastResponse(BaseModel):
    forecast: ForecastResult


# --- Bill Negotiation ---

class BillInput(BaseModel):
    name: str = Field(max_length=255)
    amount: float = Field(gt=0)
    category: str | None = None
    frequency: str | None = None


class NegotiationTip(BaseModel):
    tip: str
    typical_range: str | None = None
    potential_savings: str | None = None


class NegotiateRequest(BaseModel):
    bills: list[BillInput] = Field(max_length=10)


class NegotiateResponse(BaseModel):
    tips: list[NegotiationTip]


# --- AI Chat ---

class ChatMessage(BaseModel):
    role: Literal["user", "assistant"]
    content: str = Field(max_length=4096)


class BudgetInfo(BaseModel):
    spent: float = Field(default=0.0, ge=0)
    limit: float = Field(default=0.0, ge=0)


class ChatContext(BaseModel):
    monthly_income: float | None = None
    monthly_spending: float | None = None
    top_categories: dict[str, float] | None = None
    budget_status: dict[str, BudgetInfo] | None = None
    recent_alerts: list[str] | None = Field(default=None, max_length=10)


class ChatRequest(BaseModel):
    message: str = Field(max_length=4096)
    history: list[ChatMessage] = Field(default_factory=list, max_length=20)
    context: ChatContext | None = None


class ChatSyncResponse(BaseModel):
    content: str


# --- Health ---

class HealthResponse(BaseModel):
    status: str
    version: str
    services: dict[str, str]
