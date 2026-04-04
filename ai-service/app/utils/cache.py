"""Simple in-memory TTL cache with LRU eviction."""

import time
from collections import OrderedDict
from threading import Lock


class TTLCache:
    """Thread-safe in-memory cache with TTL and max entry limits."""

    def __init__(self, max_entries: int = 10000):
        self._cache: OrderedDict[str, tuple[float, object]] = OrderedDict()
        self._lock = Lock()
        self._max_entries = max_entries

    def get(self, key: str) -> object | None:
        with self._lock:
            if key not in self._cache:
                return None
            expires_at, value = self._cache[key]
            if time.monotonic() > expires_at:
                del self._cache[key]
                return None
            self._cache.move_to_end(key)
            return value

    def set(self, key: str, value: object, ttl: int) -> None:
        with self._lock:
            expires_at = time.monotonic() + ttl
            self._cache[key] = (expires_at, value)
            self._cache.move_to_end(key)
            while len(self._cache) > self._max_entries:
                self._cache.popitem(last=False)

    def clear(self) -> None:
        with self._lock:
            self._cache.clear()

    def size(self) -> int:
        with self._lock:
            return len(self._cache)
