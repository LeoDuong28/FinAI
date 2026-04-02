"""Bill negotiation tips powered by Gemini AI."""

import json
import logging
import re

from app.config import settings
from app.models.schemas import BillInput, NegotiationTip
from app.utils.cache import TTLCache
from app.utils.gemini_client import gemini_client

logger = logging.getLogger(__name__)

_cache = TTLCache(max_entries=1000)

# Generic tips when AI is unavailable
GENERIC_TIPS: dict[str, NegotiationTip] = {
    "utilities": NegotiationTip(
        tip="Call your provider and ask about loyalty discounts or promotional rates. Mention competitor pricing.",
        typical_range="Varies by region",
        potential_savings="10-20% off monthly bill",
    ),
    "streaming": NegotiationTip(
        tip="Consider annual plans for discounts, or rotate services monthly instead of subscribing to all.",
        typical_range="$5-$20/month per service",
        potential_savings="$50-$100/year by rotating",
    ),
    "health-insurance": NegotiationTip(
        tip="Review your plan during open enrollment. Compare marketplace plans and consider higher deductible options.",
        typical_range="Varies widely",
        potential_savings="$50-$200/month with plan adjustment",
    ),
    "gym": NegotiationTip(
        tip="Ask about corporate or student discounts. Negotiate during off-peak signup months (summer, post-New Year).",
        typical_range="$10-$60/month",
        potential_savings="$10-$20/month with negotiation",
    ),
}


async def get_negotiation_tips(
    bills: list[BillInput],
) -> list[NegotiationTip]:
    """Generate negotiation tips for bills using Gemini with fallback."""
    tips: list[NegotiationTip] = []

    for bill in bills:
        cache_key = f"neg:{bill.name.lower()}:{bill.category or 'unknown'}"
        cached = _cache.get(cache_key)
        if cached is not None:
            tips.append(cached)
            continue

        tip = await _generate_tip(bill)
        _cache.set(cache_key, tip, settings.cache_ttl_negotiation)
        tips.append(tip)

    return tips


async def _generate_tip(bill: BillInput) -> NegotiationTip:
    """Generate a single negotiation tip using Gemini or fallback."""
    # Try generic tips first
    if bill.category and bill.category in GENERIC_TIPS:
        generic = GENERIC_TIPS[bill.category]
        if not gemini_client.is_available:
            return generic

    if not gemini_client.is_available:
        return NegotiationTip(
            tip=f"Research competitive pricing for {bill.name} and call to negotiate a lower rate.",
            typical_range=None,
            potential_savings=None,
        )

    prompt = (
        f"Provide a bill negotiation tip for this bill:\n\n"
        f"Bill: {bill.name}\n"
        f"Amount: ${bill.amount:.2f}/{bill.frequency or 'month'}\n"
        f"Category: {bill.category or 'unknown'}\n\n"
        f"Respond with ONLY a JSON object:\n"
        f'{{"tip": "specific negotiation advice", '
        f'"typical_range": "price range for this type of bill", '
        f'"potential_savings": "estimated savings amount"}}'
    )

    try:
        response = await gemini_client.generate(prompt)
        match = re.search(r"\{[^{}]*\}", response)
        if match:
            data = json.loads(match.group())
            tip_text = str(data.get("tip", ""))[:1000]
            if tip_text:
                return NegotiationTip(
                    tip=tip_text,
                    typical_range=str(data.get("typical_range", ""))[:200] or None,
                    potential_savings=str(data.get("potential_savings", ""))[:200] or None,
                )
    except (json.JSONDecodeError, ValueError, TypeError) as e:
        logger.warning("Gemini negotiation tip parse error: %s", e)
    except Exception as e:
        logger.warning("Gemini negotiation tip failed: %s", e)

    if bill.category and bill.category in GENERIC_TIPS:
        return GENERIC_TIPS[bill.category]

    return NegotiationTip(
        tip=f"Research competitive pricing for {bill.name} and call to negotiate.",
        typical_range=None,
        potential_savings=None,
    )
