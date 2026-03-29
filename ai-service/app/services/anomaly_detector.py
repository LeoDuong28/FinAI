"""Anomaly detection using z-score analysis."""

import logging
import statistics
from collections import defaultdict

from app.models.schemas import AnomalyInput, AnomalyResult

logger = logging.getLogger(__name__)

Z_SCORE_THRESHOLD = 2.0
FIRST_TIME_MERCHANT_THRESHOLD = 100.0


def detect_anomalies(
    transactions: list[AnomalyInput],
    history: list[AnomalyInput],
) -> list[AnomalyResult]:
    """Detect anomalous transactions using z-score and merchant history."""
    # Build per-category statistics from history
    category_amounts: dict[str, list[float]] = defaultdict(list)
    known_merchants: set[str] = set()

    for txn in history:
        cat = txn.category or "uncategorized"
        category_amounts[cat].append(txn.amount)
        merchant = (txn.merchant_name or txn.name).lower().strip()
        if merchant:
            known_merchants.add(merchant)

    # Calculate mean and stdev per category
    category_stats: dict[str, tuple[float, float]] = {}
    for cat, amounts in category_amounts.items():
        if len(amounts) >= 3:
            mean = statistics.mean(amounts)
            stdev = statistics.stdev(amounts)
            category_stats[cat] = (mean, stdev)

    results: list[AnomalyResult] = []

    for txn in transactions:
        cat = txn.category or "uncategorized"
        merchant = (txn.merchant_name or txn.name).lower().strip()

        # Check 1: Z-score anomaly
        if cat in category_stats:
            mean, stdev = category_stats[cat]
            if stdev > 0:
                z_score = (txn.amount - mean) / stdev
                if z_score > Z_SCORE_THRESHOLD:
                    results.append(
                        AnomalyResult(
                            transaction_id=txn.id,
                            is_anomaly=True,
                            reason=f"Unusually high spending in {cat}: ${txn.amount:.2f} vs avg ${mean:.2f}",
                            z_score=round(z_score, 2),
                        )
                    )
                    continue

        # Check 2: First-time merchant with high amount
        if merchant and merchant not in known_merchants:
            if txn.amount >= FIRST_TIME_MERCHANT_THRESHOLD:
                results.append(
                    AnomalyResult(
                        transaction_id=txn.id,
                        is_anomaly=True,
                        reason=f"First-time merchant '{txn.merchant_name or txn.name}' with amount ${txn.amount:.2f}",
                        z_score=None,
                    )
                )
                continue

        results.append(
            AnomalyResult(
                transaction_id=txn.id,
                is_anomaly=False,
                reason="",
                z_score=None,
            )
        )

    return [r for r in results if r.is_anomaly]
