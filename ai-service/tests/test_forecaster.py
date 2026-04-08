"""Tests for spending forecast."""

from app.models.schemas import DailySpending
from app.services.forecaster import forecast_spending


def test_forecast_empty_history():
    result = forecast_spending([])
    assert result.predicted_total == 0.0
    assert result.daily_predictions == []


def test_forecast_basic():
    history = []
    for i in range(30):
        day = i + 1
        history.append(
            DailySpending(
                date=f"2025-01-{day:02d}",
                amount=100.0,
                category="food",
            )
        )

    result = forecast_spending(history, days=7)
    assert result.predicted_total > 0
    assert len(result.daily_predictions) == 7
    # With constant spending of $100/day, forecast should be ~$700
    assert 500 < result.predicted_total < 900


def test_forecast_with_categories():
    history = []
    for i in range(30):
        day = i + 1
        history.append(
            DailySpending(
                date=f"2025-01-{day:02d}",
                amount=60.0,
                category="food",
            )
        )
        history.append(
            DailySpending(
                date=f"2025-01-{day:02d}",
                amount=40.0,
                category="transport",
            )
        )

    result = forecast_spending(history, days=7)
    assert result.category_breakdown is not None
    assert "food" in result.category_breakdown
    assert "transport" in result.category_breakdown


def test_forecast_non_negative():
    # Decreasing spending should still produce non-negative predictions
    history = []
    for i in range(30):
        day = i + 1
        history.append(
            DailySpending(
                date=f"2025-01-{day:02d}",
                amount=max(0, 100.0 - i * 5),
            )
        )

    result = forecast_spending(history, days=30)
    for pred in result.daily_predictions:
        assert pred.amount >= 0
