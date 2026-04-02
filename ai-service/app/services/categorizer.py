"""Transaction categorization using FinBERT + Gemini dual pipeline."""

import asyncio
import json
import logging
import re
import threading

from app.config import settings
from app.models.categories import ALL_CATEGORIES, PARENT_CATEGORY, rule_based_categorize
from app.models.schemas import CategoryResult, TransactionInput
from app.utils.cache import TTLCache
from app.utils.gemini_client import gemini_client

logger = logging.getLogger(__name__)

_cache = TTLCache(max_entries=settings.cache_max_entries)

# FinBERT model (lazy loaded, protected by lock)
_classifier = None
_finbert_available = False
_finbert_lock = threading.Lock()


def _load_finbert() -> None:
    """Lazy-load FinBERT model on first use (thread-safe)."""
    global _classifier, _finbert_available
    if _classifier is not None:
        return
    with _finbert_lock:
        if _classifier is not None:
            return
        try:
            from transformers import pipeline

            _classifier = pipeline(
                "zero-shot-classification",
                model=settings.finbert_model,
                device=-1,  # CPU
            )
            _finbert_available = True
            logger.info("FinBERT model loaded successfully")
        except Exception as e:
            logger.warning("Failed to load FinBERT: %s. Using fallback.", e)
            _finbert_available = False


def _normalize_merchant(name: str, merchant_name: str | None = None) -> str:
    """Normalize merchant name for cache key and matching."""
    text = (merchant_name or name).lower().strip()
    text = re.sub(r"[^a-z0-9\s]", "", text)
    text = re.sub(r"\s+", " ", text)
    return text.strip()


def _finbert_categorize(text: str) -> tuple[str, float]:
    """Classify using FinBERT zero-shot classification."""
    if not _finbert_available or _classifier is None:
        return "uncategorized", 0.0

    result = _classifier(text, ALL_CATEGORIES, multi_label=False)
    top_label = result["labels"][0]
    top_score = result["scores"][0]
    return top_label, top_score


async def _gemini_categorize(text: str, amount: float, txn_type: str) -> tuple[str, float]:
    """Classify using Gemini API as fallback."""
    if not gemini_client.is_available:
        return "uncategorized", 0.0

    categories_str = ", ".join(ALL_CATEGORIES)
    prompt = (
        f"Categorize this financial transaction into exactly one category.\n\n"
        f"Transaction: {text}\n"
        f"Amount: ${abs(amount):.2f} ({txn_type})\n\n"
        f"Available categories: {categories_str}\n\n"
        f"Respond with ONLY a JSON object: {{\"category\": \"category-slug\", \"confidence\": 0.95}}\n"
        f"The category must be one of the available categories listed above."
    )

    try:
        response = await gemini_client.generate(prompt)
        match = re.search(r"\{[^{}]*\}", response)
        if match:
            data = json.loads(match.group())
            category = data.get("category", "uncategorized")
            confidence = float(data.get("confidence", 0.5))
            confidence = max(0.0, min(confidence, 1.0))
            if category in ALL_CATEGORIES:
                return category, confidence
    except (json.JSONDecodeError, ValueError, TypeError) as e:
        logger.warning("Gemini categorization parse error: %s", e)
    except Exception as e:
        logger.warning("Gemini categorization failed: %s", e)

    return "uncategorized", 0.0


async def categorize_transactions(
    transactions: list[TransactionInput],
) -> list[CategoryResult]:
    """Categorize transactions using the pipeline: rules -> cache -> FinBERT -> Gemini."""
    # Load FinBERT in a thread to avoid blocking the event loop
    await asyncio.to_thread(_load_finbert)
    results: list[CategoryResult] = []

    for txn in transactions:
        normalized = _normalize_merchant(txn.name, txn.merchant_name)
        cache_key = f"cat:{normalized}"

        # 1. Check cache
        cached = _cache.get(cache_key)
        if cached is not None:
            cat, conf = cached
            results.append(
                CategoryResult(
                    transaction_id=txn.id, category=cat, confidence=conf
                )
            )
            continue

        # 2. Rule-based
        rule_cat = rule_based_categorize(txn.name, txn.merchant_name)
        if rule_cat:
            _cache.set(cache_key, (rule_cat, 0.95), settings.cache_ttl_categorization)
            results.append(
                CategoryResult(
                    transaction_id=txn.id, category=rule_cat, confidence=0.95
                )
            )
            continue

        # 3. FinBERT (run in thread to avoid blocking event loop)
        text_for_model = txn.merchant_name or txn.name
        category, confidence = await asyncio.to_thread(
            _finbert_categorize, text_for_model
        )

        if confidence >= settings.finbert_confidence_threshold:
            _cache.set(
                cache_key, (category, confidence), settings.cache_ttl_categorization
            )
            results.append(
                CategoryResult(
                    transaction_id=txn.id, category=category, confidence=confidence
                )
            )
            continue

        # 4. Gemini fallback
        category, confidence = await _gemini_categorize(
            text_for_model, txn.amount, txn.type
        )
        _cache.set(
            cache_key, (category, confidence), settings.cache_ttl_categorization
        )
        results.append(
            CategoryResult(
                transaction_id=txn.id, category=category, confidence=confidence
            )
        )

    return results


def get_parent_category(slug: str) -> str | None:
    """Get the parent category for a subcategory slug."""
    return PARENT_CATEGORY.get(slug)


def is_finbert_available() -> bool:
    """Check if FinBERT model is loaded."""
    return _finbert_available
