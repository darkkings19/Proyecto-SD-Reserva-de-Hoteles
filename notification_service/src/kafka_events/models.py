from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4


RESERVATION_CONFIRMED = "ReservationConfirmed"
NOTIFICATION_SENT = "NotificationSent"
NOTIFICATION_FAILED = "NotificationFailed"
SOURCE_SERVICE = "notification-service"


@dataclass(frozen=True)
class ReservationConfirmed:
    event_id: str
    user_id: str
    reservation_id: str
    hotel_id: str
    room_type_id: str
    email: str


def should_process_event(event: dict[str, Any]) -> bool:
    return event.get("event_type") == RESERVATION_CONFIRMED


def parse_reservation_confirmed(
    event: dict[str, Any],
    default_email: str = "unknown@example.com",
) -> ReservationConfirmed:
    if event.get("event_type") != RESERVATION_CONFIRMED:
        raise ValueError(f"unsupported event_type: {event.get('event_type')}")

    payload = event.get("payload")
    if not isinstance(payload, dict):
        raise ValueError("payload must be an object")

    reservation_id = str(payload.get("reservation_id") or "").strip()
    if not reservation_id:
        raise ValueError("reservation_id is required")

    user_id = str(payload.get("user_id") or "").strip()
    if not user_id:
        raise ValueError("user_id is required")

    return ReservationConfirmed(
        event_id=str(event.get("event_id") or ""),
        user_id=user_id,
        reservation_id=reservation_id,
        hotel_id=str(payload.get("hotel_id") or ""),
        room_type_id=str(payload.get("room_type_id") or ""),
        email=str(payload.get("email") or default_email),
    )


def idempotency_key(reservation_id: str) -> str:
    return f"reservation-confirmation:{reservation_id}"


def build_notification_sent_event(
    reservation: ReservationConfirmed,
    correlation_id: str | None = None,
) -> dict[str, Any]:
    return _event_envelope(
        event_type=NOTIFICATION_SENT,
        correlation_id=correlation_id or reservation.event_id,
        payload={
            "user_id": reservation.user_id,
            "reservation_id": reservation.reservation_id,
            "notification_type": "CONFIRMACION",
            "channel": "EMAIL",
            "status": "SENT",
        },
    )


def build_notification_failed_event(
    reservation: ReservationConfirmed | None,
    reason: str,
    correlation_id: str | None = None,
) -> dict[str, Any]:
    payload = {
        "notification_type": "CONFIRMACION",
        "channel": "EMAIL",
        "status": "FAILED",
        "reason": reason,
    }
    if reservation is not None:
        payload.update(
            {
                "user_id": reservation.user_id,
                "reservation_id": reservation.reservation_id,
            }
        )

    return _event_envelope(
        event_type=NOTIFICATION_FAILED,
        correlation_id=correlation_id or (reservation.event_id if reservation else None),
        payload=payload,
    )


def _event_envelope(
    event_type: str,
    payload: dict[str, Any],
    correlation_id: str | None = None,
) -> dict[str, Any]:
    event = {
        "event_id": str(uuid4()),
        "event_type": event_type,
        "version": 1,
        "source_service": SOURCE_SERVICE,
        "occurred_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "payload": payload,
    }
    if correlation_id:
        event["correlation_id"] = correlation_id
    return event
