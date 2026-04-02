"""Spending forecast using linear regression."""

import logging
from collections import defaultdict
from datetime import datetime, timedelta

import numpy as np
from sklearn.linear_model import LinearRegression

from app.models.schemas import DailySpending, ForecastResult

logger = logging.getLogger(__name__)


def _rolling_average(values: list[float], window: int = 7) -> list[float]:
    """Compute rolling average to smooth daily spending."""
    if len(values) <= window:
        avg = sum(values) / len(values) if values else 0
        return [avg] * len(values)

    result = []
    for i in range(len(values)):
        start = max(0, i - window + 1)
        window_vals = values[start : i + 1]
        result.append(sum(window_vals) / len(window_vals))
    return result


def forecast_spending(
    history: list[DailySpending], days: int = 30
) -> ForecastResult:
    """Forecast spending for the next N days using linear regression."""
    if not history:
        return ForecastResult(
            predicted_total=0.0,
            daily_predictions=[],
            category_breakdown=None,
        )

    # Parse dates and aggregate by day
    daily_totals: dict[str, float] = defaultdict(float)
    category_totals: dict[str, float] = defaultdict(float)

    for entry in history:
        daily_totals[entry.date] += entry.amount
        if entry.category:
            category_totals[entry.category] += entry.amount

    if not daily_totals:
        return ForecastResult(
            predicted_total=0.0, daily_predictions=[], category_breakdown=None
        )

    # Sort by date
    sorted_dates = sorted(daily_totals.keys())
    amounts = [daily_totals[d] for d in sorted_dates]

    # Apply 7-day rolling average
    smoothed = _rolling_average(amounts)

    # Prepare features for linear regression
    x = np.arange(len(smoothed)).reshape(-1, 1)
    y = np.array(smoothed)

    model = LinearRegression()
    model.fit(x, y)

    # Predict next N days
    last_date = datetime.strptime(sorted_dates[-1], "%Y-%m-%d")
    future_x = np.arange(len(smoothed), len(smoothed) + days).reshape(-1, 1)
    predictions = model.predict(future_x)

    # Ensure predictions are non-negative
    predictions = np.maximum(predictions, 0)

    daily_predictions: list[DailySpending] = []
    predicted_total = 0.0

    for i, pred in enumerate(predictions):
        pred_date = last_date + timedelta(days=i + 1)
        amount = round(float(pred), 2)
        predicted_total += amount
        daily_predictions.append(
            DailySpending(
                date=pred_date.strftime("%Y-%m-%d"),
                amount=amount,
            )
        )

    # Category breakdown: proportional to historical spending
    total_historical = sum(category_totals.values())
    category_breakdown = None
    if total_historical > 0:
        category_breakdown = {
            cat: round(predicted_total * (amt / total_historical), 2)
            for cat, amt in category_totals.items()
        }

    return ForecastResult(
        predicted_total=round(predicted_total, 2),
        daily_predictions=daily_predictions,
        category_breakdown=category_breakdown,
    )
