import asyncio
import inspect
import logging
import random
import uuid

import grpc

# Timeouts por tipo de llamada (segundos). El de creacion de reserva es mayor a
# la suma de los timeouts internos de la SAGA en reservation_service hacia
# usuario e inventario (5s + 5s), para no cortar una operacion que todavia
# puede completarse del lado del servidor.
TIMEOUT_USER_CALL = 3.0
TIMEOUT_SEARCH = 3.0
TIMEOUT_RESERVATION_READ = 4.0
TIMEOUT_RESERVATION_CREATE = 12.0

# Solo se reintentan errores de transporte transitorios. Errores de negocio
# (NOT_FOUND, UNAUTHENTICATED, RESOURCE_EXHAUSTED, etc.) nunca se reintentan:
# reintentar "no hay stock" o "credenciales invalidas" no cambia el resultado.
RETRYABLE_CODES = {grpc.StatusCode.UNAVAILABLE, grpc.StatusCode.DEADLINE_EXCEEDED}

_MAX_ATTEMPTS = 3
_BASE_DELAY_SECONDS = 0.25


def new_idempotency_key() -> str:
    return uuid.uuid4().hex


async def call_with_retry(func, *args, operation: str = "", **kwargs):
    """Ejecuta func (sync o coroutine) reintentando con backoff exponencial +
    jitter ante errores de transporte transitorios. Reusa la misma logica para
    los clientes sync (users_client, inventory_client) y async (reservations_client)."""
    attempt = 0
    while True:
        attempt += 1
        try:
            result = func(*args, **kwargs)
            if inspect.isawaitable(result):
                result = await result
            return result
        except grpc.RpcError as exc:
            code = exc.code() if hasattr(exc, "code") else None
            if code not in RETRYABLE_CODES or attempt >= _MAX_ATTEMPTS:
                raise
            delay = _BASE_DELAY_SECONDS * (2 ** (attempt - 1)) + random.uniform(0, 0.1)
            logging.warning(
                "retry attempt=%d operation=%s reason=%s delay=%.2fs",
                attempt, operation, code, delay,
            )
            await asyncio.sleep(delay)
