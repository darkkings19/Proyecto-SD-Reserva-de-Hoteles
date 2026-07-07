import pytest
from unittest.mock import Mock

from kafka_events.models import (
    NOTIFICATION_FAILED,
    NOTIFICATION_SENT,
    RESERVATION_CONFIRMED,
    build_notification_failed_event,
    build_notification_sent_event,
    idempotency_key,
    parse_reservation_confirmed,
    should_process_event,
)
from kafka_events.worker import (
    CommitFailedError,
    KafkaSettings,
    NoBrokersAvailable,
    NotificationKafkaWorker,
)


def test_parse_valid_reservation_confirmed():
    reservation = parse_reservation_confirmed(valid_reservation_confirmed_event())

    assert reservation.event_id == "evt-1"
    assert reservation.user_id == "usr-1001"
    assert reservation.reservation_id == "res-1001"
    assert reservation.hotel_id == "hotel-001"
    assert reservation.room_type_id == "room-std"


def test_ignore_reservation_created():
    event = valid_reservation_confirmed_event()
    event["event_type"] = "ReservationCreated"

    assert should_process_event(event) is False


def test_reject_reservation_confirmed_without_reservation_id():
    event = valid_reservation_confirmed_event()
    event["payload"]["reservation_id"] = ""

    with pytest.raises(ValueError, match="reservation_id is required"):
        parse_reservation_confirmed(event)


def test_build_notification_sent():
    reservation = parse_reservation_confirmed(valid_reservation_confirmed_event())
    event = build_notification_sent_event(reservation, correlation_id="evt-1")

    assert event["event_type"] == NOTIFICATION_SENT
    assert event["version"] == 1
    assert event["source_service"] == "notification-service"
    assert event["correlation_id"] == "evt-1"
    assert event["payload"]["reservation_id"] == "res-1001"
    assert event["payload"]["status"] == "SENT"


def test_build_notification_failed():
    reservation = parse_reservation_confirmed(valid_reservation_confirmed_event())
    event = build_notification_failed_event(reservation, reason="provider unavailable")

    assert event["event_type"] == NOTIFICATION_FAILED
    assert event["version"] == 1
    assert event["source_service"] == "notification-service"
    assert event["payload"]["reservation_id"] == "res-1001"
    assert event["payload"]["status"] == "FAILED"
    assert event["payload"]["reason"] == "provider unavailable"


def test_idempotency_key_uses_reservation_id():
    assert idempotency_key("res-1001") == "reservation-confirmation:res-1001"


def test_worker_does_not_break_when_kafka_is_unavailable():
    worker = NotificationKafkaWorker(make_settings(), Mock())
    attempts = {"count": 0}

    def fail_connect():
        attempts["count"] += 1
        worker._stop_event.set()
        raise NoBrokersAvailable("no brokers")

    worker._connect_clients = fail_connect
    worker._run()

    assert attempts["count"] == 1


def test_duplicate_event_is_ignored_without_publishing():
    servicer = Mock()
    servicer.process_confirmation.return_value = False
    worker = NotificationKafkaWorker(make_settings(), servicer)
    worker._publish_notification_event = Mock()

    result = worker.handle_event(valid_reservation_confirmed_event())

    assert result == "duplicate"
    worker._publish_notification_event.assert_not_called()


def test_commit_error_is_handled_without_breaking():
    class FailingConsumer:
        def commit(self):
            raise CommitFailedError("commit failed")

    worker = NotificationKafkaWorker(make_settings(), Mock())
    worker._consumer = FailingConsumer()

    worker._commit_message()


def make_settings():
    return KafkaSettings(
        enabled=True,
        bootstrap_servers="kafka:9092",
        reservations_topic="origenx.reservations.events",
        notifications_topic="origenx.notifications.events",
        consumer_group="notification-service",
        client_id="notification-service",
        default_email="unknown@example.com",
    )


def valid_reservation_confirmed_event():
    return {
        "event_id": "evt-1",
        "event_type": RESERVATION_CONFIRMED,
        "version": 1,
        "source_service": "reservation-service",
        "occurred_at": "2026-07-05T16:00:00Z",
        "payload": {
            "reservation_id": "res-1001",
            "user_id": "usr-1001",
            "hotel_id": "hotel-001",
            "room_type_id": "room-std",
            "status": "CONFIRMED",
            "monto_total": 150.5,
        },
    }
