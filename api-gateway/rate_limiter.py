import hashlib
import math
import time
from collections import deque
from typing import Deque, Dict, Tuple

from fastapi import Request
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware


def client_key(request: Request) -> str:
    """Identifica al cliente por su token de sesion si lo envia, o por IP si no."""
    auth = request.headers.get("authorization")
    if auth:
        return "token:" + hashlib.sha256(auth.encode()).hexdigest()[:16]
    client = request.client
    return "ip:" + (client.host if client else "unknown")


class TokenBucketLimiter:
    """Token bucket: permite rafagas cortas (hasta 'capacity') y limita el
    promedio sostenido a 'refill_per_second'. Pensado para trafico legitimo
    que incluye reintentos automaticos del propio gateway."""

    def __init__(self, capacity: int, refill_per_second: float):
        self.capacity = float(capacity)
        self.refill_per_second = refill_per_second
        self._tokens: Dict[str, float] = {}
        self._last_refill: Dict[str, float] = {}

    def allow(self, key: str) -> Tuple[bool, float]:
        now = time.monotonic()
        tokens = self._tokens.get(key, self.capacity)
        last = self._last_refill.get(key, now)

        elapsed = now - last
        tokens = min(self.capacity, tokens + elapsed * self.refill_per_second)

        if tokens >= 1:
            tokens -= 1
            self._tokens[key] = tokens
            self._last_refill[key] = now
            return True, 0.0

        self._tokens[key] = tokens
        self._last_refill[key] = now
        retry_after = (1 - tokens) / self.refill_per_second
        return False, retry_after


class SlidingWindowLimiter:
    """Sliding window log: no perdona rafagas dentro de la ventana. Pensado
    para endpoints donde una rafaga en si misma es el ataque (fuerza bruta)."""

    def __init__(self, max_events: int, window_seconds: float):
        self.max_events = max_events
        self.window_seconds = window_seconds
        self._events: Dict[str, Deque[float]] = {}

    def allow(self, key: str) -> Tuple[bool, float]:
        now = time.monotonic()
        window_start = now - self.window_seconds
        events = self._events.setdefault(key, deque())

        while events and events[0] < window_start:
            events.popleft()

        if len(events) >= self.max_events:
            retry_after = events[0] + self.window_seconds - now
            return False, max(retry_after, 0.0)

        events.append(now)
        return True, 0.0


# Cupo global: 20 peticiones de rafaga, recarga sostenida de 10/s por cliente.
global_limiter = TokenBucketLimiter(capacity=20, refill_per_second=10.0)

# Especifico para /login: maximo 5 intentos por 60s por combinacion email+IP.
login_limiter = SlidingWindowLimiter(max_events=5, window_seconds=60.0)


def _retry_after_response(retry_after: float, message: str) -> JSONResponse:
    seconds = max(1, math.ceil(retry_after))
    return JSONResponse(
        status_code=429,
        content={"detail": message},
        headers={"Retry-After": str(seconds)},
    )


class RateLimitMiddleware(BaseHTTPMiddleware):
    """Aplica el token bucket global a toda peticion HTTP excepto /health."""

    async def dispatch(self, request: Request, call_next):
        if request.url.path == "/health":
            return await call_next(request)

        key = client_key(request)
        allowed, retry_after = global_limiter.allow(key)
        if not allowed:
            return _retry_after_response(retry_after, "Demasiadas solicitudes, intenta de nuevo mas tarde")

        return await call_next(request)


def check_login_rate_limit(request: Request, email: str) -> None:
    """Sliding window especifico de /login, keyed por email+IP. Se llama desde
    el propio endpoint (ya con el body parseado) para no tener que leer el
    stream de la peticion dos veces en middleware."""
    client = request.client
    ip = client.host if client else "unknown"
    key = f"{email.lower()}:{ip}"
    allowed, retry_after = login_limiter.allow(key)
    if not allowed:
        raise RateLimitExceeded(retry_after)


class RateLimitExceeded(Exception):
    def __init__(self, retry_after: float):
        self.retry_after = retry_after
        super().__init__("rate limit exceeded")
