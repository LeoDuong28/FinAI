"""Tests for transaction categorization."""

import pytest

from app.models.categories import rule_based_categorize
from app.models.schemas import TransactionInput
from app.services.categorizer import _normalize_merchant


def test_rule_based_starbucks():
    result = rule_based_categorize("STARBUCKS #1234", None)
    assert result == "coffee"


def test_rule_based_walmart():
    result = rule_based_categorize("WAL-MART", "Walmart Supercenter")
    assert result == "groceries"


def test_rule_based_netflix():
    result = rule_based_categorize("NETFLIX.COM", None)
    assert result == "streaming"


def test_rule_based_uber():
    result = rule_based_categorize("UBER TRIP", None)
    assert result == "ride-share"


def test_rule_based_unknown():
    result = rule_based_categorize("RANDOM SHOP 42", None)
    assert result is None


def test_rule_based_transfer():
    result = rule_based_categorize("Venmo Payment", None)
    assert result == "transfers"


def test_rule_based_salary():
    result = rule_based_categorize("DIRECT DEPOSIT - ACME CORP", None)
    assert result == "salary"


def test_normalize_merchant():
    assert _normalize_merchant("STARBUCKS #1234") == "starbucks 1234"
    assert _normalize_merchant("  McDonald's  ") == "mcdonalds"
    assert _normalize_merchant("WAL-MART", "Walmart") == "walmart"


def test_normalize_merchant_empty():
    assert _normalize_merchant("", None) == ""
