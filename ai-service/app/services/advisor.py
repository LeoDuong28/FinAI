"""AI financial advisor for chat conversations."""

import logging

from app.models.schemas import ChatContext, ChatMessage
from app.utils.gemini_client import gemini_client

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = """You are a friendly, knowledgeable personal finance advisor. You help users understand their spending, make better financial decisions, and achieve their financial goals.

Guidelines:
- Be concise and actionable. Give specific advice, not generic platitudes.
- When discussing money, use dollar amounts and percentages.
- If the user asks about investing, remind them you're not a licensed financial advisor.
- Reference their actual financial data when available.
- Keep responses under 200 words unless the user asks for detailed analysis.
- Be encouraging but honest about overspending.
- Suggest specific, practical steps the user can take.

{context}"""


def _build_context_string(context: ChatContext | None) -> str:
    """Build a context string from the user's financial data."""
    if context is None:
        return "No financial data available for this user."

    parts = []
    if context.monthly_income is not None:
        parts.append(f"Monthly income: ${context.monthly_income:,.2f}")
    if context.monthly_spending is not None:
        parts.append(f"Monthly spending: ${context.monthly_spending:,.2f}")
    if context.top_categories:
        cats = ", ".join(
            f"{k}: ${v:,.2f}" for k, v in sorted(
                context.top_categories.items(), key=lambda x: -x[1]
            )[:5]
        )
        parts.append(f"Top spending categories: {cats}")
    if context.budget_status:
        budgets = []
        for name, info in context.budget_status.items():
            pct = (info.spent / info.limit * 100) if info.limit > 0 else 0
            budgets.append(f"{name}: ${info.spent:,.2f}/${info.limit:,.2f} ({pct:.0f}%)")
        parts.append(f"Budget status: {'; '.join(budgets)}")
    if context.recent_alerts:
        parts.append(f"Recent alerts: {'; '.join(context.recent_alerts[:3])}")

    if not parts:
        return "No financial data available for this user."
    return "User's financial context:\n" + "\n".join(parts)


def _sanitize_content(text: str) -> str:
    """Strip control characters and role markers to mitigate prompt injection."""
    import re
    # Remove carriage returns and excessive newlines
    text = text.replace("\r", "")
    # Collapse multiple newlines into one
    text = re.sub(r"\n{2,}", "\n", text)
    # Strip role markers that could confuse the model
    text = re.sub(r"(?i)^(User|Assistant|System)\s*:", "", text, flags=re.MULTILINE)
    return text.strip()


def _build_prompt(
    message: str,
    history: list[ChatMessage],
    context: ChatContext | None,
) -> str:
    """Build the full prompt for Gemini."""
    context_str = _build_context_string(context)
    system = SYSTEM_PROMPT.format(context=context_str)

    parts = [system, ""]
    for msg in history[-10:]:  # Last 10 messages for context
        role = "User" if msg.role == "user" else "Assistant"
        parts.append(f"{role}: {_sanitize_content(msg.content)}")

    parts.append(f"User: {_sanitize_content(message)}")
    parts.append("Assistant:")

    return "\n".join(parts)


async def chat(
    message: str,
    history: list[ChatMessage],
    context: ChatContext | None,
) -> str:
    """Generate a chat response using Gemini."""
    if not gemini_client.is_available:
        return "I'm sorry, the AI advisor is temporarily unavailable. Please try again later."

    prompt = _build_prompt(message, history, context)

    try:
        response = await gemini_client.generate(prompt)
        return response.strip()
    except Exception as e:
        logger.error("Chat generation failed: %s", e)
        return "I'm having trouble generating a response right now. Please try again in a moment."


async def chat_stream(
    message: str,
    history: list[ChatMessage],
    context: ChatContext | None,
):
    """Generate a streaming chat response using Gemini."""
    if not gemini_client.is_available:
        yield "I'm sorry, the AI advisor is temporarily unavailable. Please try again later."
        return

    prompt = _build_prompt(message, history, context)

    try:
        async for chunk in gemini_client.generate_stream(prompt):
            yield chunk
    except Exception as e:
        logger.error("Chat stream failed: %s", e)
        yield "I'm having trouble generating a response right now. Please try again."
