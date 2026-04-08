"""Tests for subscription detection."""

from app.models.schemas import TransactionHistory
from app.services.subscription_detector import detect_subscriptions


def _make_monthly_txns(
    merchant: str, amount: float, months: int = 6
) -> list[TransactionHistory]:
    """Generate monthly transactions for testing."""
    txns = []
    for i in range(months):
        month = 12 - i if 12 - i > 0 else 12 - i + 12
        year = 2025 if month <= 12 else 2024
        if month > 12:
            month -= 12
        txns.append(
            TransactionHistory(
                name=merchant,
                merchant_name=merchant,
                amount=amount,
                date=f"{year}-{month:02d}-15",
                type="debit",
            )
        )
    return txns


def test_detect_monthly_subscription():
    txns = _make_monthly_txns("Netflix", 15.99)
    subs = detect_subscriptions(txns)
    assert len(subs) == 1
    assert subs[0].merchant_name == "Netflix"
    assert subs[0].amount == 15.99
    assert subs[0].frequency == "monthly"
    assert subs[0].confidence > 0.5


def test_no_subscription_with_few_transactions():
    txns = [
        TransactionHistory(
            name="Random Shop",
            amount=25.00,
            date="2025-01-15",
            type="debit",
        ),
        TransactionHistory(
            name="Random Shop",
            amount=30.00,
            date="2025-02-15",
            type="debit",
        ),
    ]
    subs = detect_subscriptions(txns)
    assert len(subs) == 0


def test_no_subscription_with_variable_amounts():
    txns = []
    for i in range(6):
        txns.append(
            TransactionHistory(
                name="Grocery Store",
                merchant_name="Grocery Store",
                amount=50.0 + i * 30,  # highly variable
                date=f"2025-{i + 1:02d}-15",
                type="debit",
            )
        )
    subs = detect_subscriptions(txns)
    assert len(subs) == 0


def test_multiple_subscriptions():
    txns = _make_monthly_txns("Netflix", 15.99) + _make_monthly_txns(
        "Spotify", 9.99
    )
    subs = detect_subscriptions(txns)
    assert len(subs) == 2
    names = {s.merchant_name for s in subs}
    assert "Netflix" in names
    assert "Spotify" in names


def test_ignores_credit_transactions():
    txns = []
    for i in range(6):
        txns.append(
            TransactionHistory(
                name="Refund",
                amount=15.99,
                date=f"2025-{i + 1:02d}-15",
                type="credit",
            )
        )
    subs = detect_subscriptions(txns)
    assert len(subs) == 0
