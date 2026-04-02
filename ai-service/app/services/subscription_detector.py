"""Subscription detection using transaction pattern analysis."""

import logging
import statistics
from collections import defaultdict
from datetime import datetime, timedelta

from Levenshtein import ratio as levenshtein_ratio

from app.models.schemas import DetectedSubscription, TransactionHistory

logger = logging.getLogger(__name__)

# Frequency detection parameters
FREQUENCY_MAP = {
    "weekly": (5, 9),
    "monthly": (26, 35),
    "quarterly": (80, 100),
    "yearly": (350, 380),
}
MIN_OCCURRENCES = 3
MAX_CV = 0.15  # coefficient of variation threshold for amount consistency
MERCHANT_SIMILARITY_THRESHOLD = 0.85


def _normalize_name(name: str) -> str:
    return (name or "").lower().strip()


def _group_by_merchant(
    transactions: list[TransactionHistory],
) -> dict[str, list[TransactionHistory]]:
    """Group transactions by normalized merchant name using fuzzy matching."""
    groups: dict[str, list[TransactionHistory]] = defaultdict(list)
    canonical_names: dict[str, str] = {}

    for txn in transactions:
        if txn.type != "debit":
            continue
        name = _normalize_name(txn.merchant_name or txn.name)
        if not name:
            continue

        # Find matching canonical name
        matched = False
        for canonical in canonical_names:
            if levenshtein_ratio(name, canonical) >= MERCHANT_SIMILARITY_THRESHOLD:
                groups[canonical].append(txn)
                matched = True
                break

        if not matched:
            canonical_names[name] = name
            groups[name].append(txn)

    return groups


def _detect_frequency(dates: list[datetime]) -> tuple[str | None, float]:
    """Detect billing frequency from transaction dates.

    Returns (frequency, confidence) or (None, 0.0).
    """
    if len(dates) < MIN_OCCURRENCES:
        return None, 0.0

    sorted_dates = sorted(dates)
    intervals = [
        (sorted_dates[i + 1] - sorted_dates[i]).days
        for i in range(len(sorted_dates) - 1)
    ]

    if not intervals:
        return None, 0.0

    median_interval = statistics.median(intervals)

    for frequency, (low, high) in FREQUENCY_MAP.items():
        if low <= median_interval <= high:
            # Calculate consistency score
            deviations = [abs(iv - median_interval) for iv in intervals]
            mean_deviation = statistics.mean(deviations) if deviations else 0
            tolerance = (high - low) / 2
            consistency = max(0.0, 1.0 - (mean_deviation / tolerance))
            confidence = min(1.0, consistency * 0.8 + 0.2)
            return frequency, confidence

    return None, 0.0


def _predict_next_billing(
    last_date: datetime, frequency: str
) -> str | None:
    """Predict the next billing date based on frequency."""
    interval_days = {
        "weekly": 7,
        "monthly": 30,
        "quarterly": 90,
        "yearly": 365,
    }
    days = interval_days.get(frequency)
    if days is None:
        return None
    next_date = last_date + timedelta(days=days)
    return next_date.strftime("%Y-%m-%d")


def detect_subscriptions(
    transactions: list[TransactionHistory],
) -> list[DetectedSubscription]:
    """Detect recurring subscriptions from transaction history."""
    groups = _group_by_merchant(transactions)
    subscriptions: list[DetectedSubscription] = []

    for merchant, txns in groups.items():
        if len(txns) < MIN_OCCURRENCES:
            continue

        amounts = [txn.amount for txn in txns]
        mean_amount = statistics.mean(amounts)

        # Check amount consistency (coefficient of variation)
        if mean_amount > 0 and len(amounts) > 1:
            stdev = statistics.stdev(amounts)
            cv = stdev / mean_amount
            if cv > MAX_CV:
                continue
        elif mean_amount <= 0:
            continue

        # Parse dates
        dates: list[datetime] = []
        for txn in txns:
            try:
                dates.append(datetime.strptime(txn.date, "%Y-%m-%d"))
            except ValueError:
                continue

        if len(dates) < MIN_OCCURRENCES:
            continue

        frequency, confidence = _detect_frequency(dates)
        if frequency is None:
            continue

        sorted_dates = sorted(dates)
        last_charged = sorted_dates[-1].strftime("%Y-%m-%d")
        next_billing = _predict_next_billing(sorted_dates[-1], frequency)

        # Use the original merchant name (capitalized) from most recent txn
        display_name = txns[-1].merchant_name or txns[-1].name

        subscriptions.append(
            DetectedSubscription(
                merchant_name=display_name,
                amount=round(mean_amount, 2),
                frequency=frequency,
                confidence=round(confidence, 2),
                next_billing=next_billing,
                last_charged=last_charged,
                transaction_count=len(txns),
            )
        )

    # Sort by confidence descending
    subscriptions.sort(key=lambda s: s.confidence, reverse=True)
    return subscriptions
