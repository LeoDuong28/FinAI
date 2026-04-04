"""Google Gemini API client with rate limiting and retry."""

import asyncio
import logging
import threading
import time

from app.config import settings

logger = logging.getLogger(__name__)


class GeminiClient:
    """Wrapper around Google Gemini API with rate limiting.

    Thread-safe and async-safe via locks.
    """

    def __init__(self) -> None:
        self._model = None
        self._last_request_time: float = 0
        self._min_interval: float = 60.0 / max(settings.gemini_max_rpm, 1)
        self._initialized = False
        self._init_lock = threading.Lock()
        self._rate_lock: asyncio.Lock | None = None

    def _ensure_initialized(self) -> None:
        if self._initialized:
            return
        with self._init_lock:
            if self._initialized:
                return
            if not settings.gemini_api_key:
                raise RuntimeError("GEMINI_API_KEY not configured")

            import google.generativeai as genai

            genai.configure(api_key=settings.gemini_api_key)
            self._model = genai.GenerativeModel(settings.gemini_model)
            self._initialized = True

    async def _wait_rate_limit(self) -> None:
        """Wait for rate limit window under lock to prevent concurrent bursts."""
        if self._rate_lock is None:
            self._rate_lock = asyncio.Lock()
        async with self._rate_lock:
            now = time.monotonic()
            elapsed = now - self._last_request_time
            if elapsed < self._min_interval:
                await asyncio.sleep(self._min_interval - elapsed)
            self._last_request_time = time.monotonic()

    async def generate(self, prompt: str, max_retries: int = 3) -> str:
        """Generate text with rate limiting and exponential backoff retry."""
        self._ensure_initialized()
        if self._model is None:
            raise RuntimeError("Gemini model not initialized")

        for attempt in range(max_retries):
            await self._wait_rate_limit()

            try:
                response = await asyncio.wait_for(
                    asyncio.to_thread(self._model.generate_content, prompt),
                    timeout=30.0,
                )
                if response.text:
                    return response.text
                return ""
            except Exception as e:
                delay = (2**attempt) * 1.0
                logger.warning(
                    "Gemini API error (attempt %d/%d): %s, retrying in %.1fs",
                    attempt + 1,
                    max_retries,
                    str(e),
                    delay,
                )
                if attempt < max_retries - 1:
                    await asyncio.sleep(delay)
                else:
                    raise

        return ""

    async def generate_stream(self, prompt: str, timeout: float = 60.0):
        """Generate text with streaming response and timeout protection."""
        self._ensure_initialized()
        if self._model is None:
            raise RuntimeError("Gemini model not initialized")

        await self._wait_rate_limit()

        response = await asyncio.to_thread(
            self._model.generate_content, prompt, stream=True
        )

        deadline = time.monotonic() + timeout
        for chunk in response:
            if time.monotonic() > deadline:
                logger.warning("Gemini streaming response timed out after %.0fs", timeout)
                return
            if chunk.text:
                yield chunk.text

    @property
    def is_available(self) -> bool:
        return bool(settings.gemini_api_key)


gemini_client = GeminiClient()
