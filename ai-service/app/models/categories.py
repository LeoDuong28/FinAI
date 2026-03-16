"""Finance category taxonomy for transaction categorization."""

CATEGORIES: dict[str, list[str]] = {
    "income": ["salary", "freelance", "investments", "refunds", "other-income"],
    "housing": ["rent-mortgage", "utilities", "home-insurance", "maintenance"],
    "transportation": ["gas", "public-transit", "ride-share", "parking", "car-payment"],
    "food": ["groceries", "restaurants", "coffee", "delivery"],
    "shopping": ["clothing", "electronics", "home-goods", "personal-care"],
    "entertainment": ["streaming", "gaming", "events", "hobbies"],
    "health": ["medical", "pharmacy", "gym", "health-insurance"],
    "financial": ["bank-fees", "interest", "transfers"],
    "education": ["tuition", "books", "courses"],
    "travel": ["flights", "hotels", "vacation"],
}

# Flat list of all category slugs
ALL_CATEGORIES: list[str] = []
for subcategories in CATEGORIES.values():
    ALL_CATEGORIES.extend(subcategories)

# Parent category lookup
PARENT_CATEGORY: dict[str, str] = {}
for parent, children in CATEGORIES.items():
    for child in children:
        PARENT_CATEGORY[child] = parent

# Keyword-based rules for common merchants (fallback)
MERCHANT_RULES: dict[str, str] = {
    "walmart": "groceries",
    "target": "groceries",
    "costco": "groceries",
    "kroger": "groceries",
    "whole foods": "groceries",
    "trader joe": "groceries",
    "aldi": "groceries",
    "safeway": "groceries",
    "publix": "groceries",
    "starbucks": "coffee",
    "dunkin": "coffee",
    "mcdonald": "restaurants",
    "chipotle": "restaurants",
    "subway": "restaurants",
    "chick-fil-a": "restaurants",
    "wendy": "restaurants",
    "burger king": "restaurants",
    "taco bell": "restaurants",
    "pizza hut": "restaurants",
    "domino": "restaurants",
    "doordash": "delivery",
    "uber eats": "delivery",
    "grubhub": "delivery",
    "instacart": "groceries",
    "uber": "ride-share",
    "lyft": "ride-share",
    "shell": "gas",
    "chevron": "gas",
    "exxon": "gas",
    "bp ": "gas",
    "netflix": "streaming",
    "spotify": "streaming",
    "hulu": "streaming",
    "disney+": "streaming",
    "disney plus": "streaming",
    "apple music": "streaming",
    "youtube": "streaming",
    "hbo": "streaming",
    "amazon prime": "streaming",
    "amazon": "shopping",
    "best buy": "electronics",
    "apple.com": "electronics",
    "apple store": "electronics",
    "nike": "clothing",
    "adidas": "clothing",
    "zara": "clothing",
    "h&m": "clothing",
    "old navy": "clothing",
    "gap": "clothing",
    "tj maxx": "clothing",
    "ross": "clothing",
    "nordstrom": "clothing",
    "macy": "clothing",
    "planet fitness": "gym",
    "anytime fitness": "gym",
    "la fitness": "gym",
    "equinox": "gym",
    "cvs": "pharmacy",
    "walgreens": "pharmacy",
    "rite aid": "pharmacy",
    "comcast": "utilities",
    "xfinity": "utilities",
    "verizon": "utilities",
    "at&t": "utilities",
    "t-mobile": "utilities",
    "sprint": "utilities",
    "electric": "utilities",
    "water bill": "utilities",
    "gas bill": "utilities",
    "geico": "home-insurance",
    "state farm": "home-insurance",
    "allstate": "home-insurance",
    "progressive": "home-insurance",
    "venmo": "transfers",
    "zelle": "transfers",
    "paypal": "transfers",
    "cash app": "transfers",
    "atm": "transfers",
    "transfer": "transfers",
    "direct deposit": "salary",
    "payroll": "salary",
    "deposit": "salary",
    "interest earned": "interest",
    "interest charge": "interest",
    "overdraft": "bank-fees",
    "monthly fee": "bank-fees",
    "late fee": "bank-fees",
    "annual fee": "bank-fees",
}


def rule_based_categorize(name: str, merchant_name: str | None = None) -> str | None:
    """Attempt to categorize using keyword rules. Returns category slug or None."""
    text = (merchant_name or name).lower().strip()
    for keyword, category in MERCHANT_RULES.items():
        if keyword in text:
            return category
    return None
