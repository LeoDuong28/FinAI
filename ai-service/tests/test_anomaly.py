"""Tests for anomaly detection."""

from app.models.schemas import AnomalyInput
from app.services.anomaly_detector import detect_anomalies


def test_detect_high_spending_anomaly():
    history = [
        AnomalyInput(
            id=f"h{i}", name="Grocery", amount=45.0 + i * 0.5, date=f"2025-01-{i + 1:02d}", category="food"
        )
        for i in range(20)
    ]

    transactions = [
        AnomalyInput(
            id="t1", name="Grocery", amount=500.0, date="2025-02-01", category="food"
        )
    ]

    anomalies = detect_anomalies(transactions, history)
    assert len(anomalies) == 1
    assert anomalies[0].transaction_id == "t1"
    assert anomalies[0].is_anomaly is True
    assert anomalies[0].z_score is not None
    assert anomalies[0].z_score > 2.0


def test_no_anomaly_for_normal_spending():
    history = [
        AnomalyInput(
            id=f"h{i}", name="Grocery", amount=50.0 + i, date=f"2025-01-{i + 1:02d}", category="food"
        )
        for i in range(20)
    ]

    transactions = [
        AnomalyInput(
            id="t1", name="Grocery", amount=55.0, date="2025-02-01", category="food"
        )
    ]

    anomalies = detect_anomalies(transactions, history)
    assert len(anomalies) == 0


def test_first_time_merchant_high_amount():
    history = [
        AnomalyInput(
            id=f"h{i}", name="Known Store", amount=30.0, date=f"2025-01-{i + 1:02d}", category="shopping"
        )
        for i in range(10)
    ]

    transactions = [
        AnomalyInput(
            id="t1", name="New Fancy Store", merchant_name="New Fancy Store",
            amount=250.0, date="2025-02-01", category="shopping"
        )
    ]

    anomalies = detect_anomalies(transactions, history)
    assert len(anomalies) == 1
    assert "First-time" in anomalies[0].reason


def test_first_time_merchant_low_amount_ok():
    history = [
        AnomalyInput(
            id=f"h{i}", name="Known Store", amount=30.0, date=f"2025-01-{i + 1:02d}", category="shopping"
        )
        for i in range(10)
    ]

    transactions = [
        AnomalyInput(
            id="t1", name="New Small Shop", merchant_name="New Small Shop",
            amount=15.0, date="2025-02-01", category="shopping"
        )
    ]

    anomalies = detect_anomalies(transactions, history)
    assert len(anomalies) == 0
