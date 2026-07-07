from kafka_events import (
    gateway_inventory_search_requested,
    gateway_reservation_requested,
    gateway_user_logged_in,
    gateway_user_registered,
)


def assert_common(event, event_type):
    assert event["event_id"]
    assert event["event_type"] == event_type
    assert event["version"] == 1
    assert event["source_service"] == "api-gateway"
    assert event["occurred_at"]
    assert "payload" in event


def assert_no_sensitive_data(event):
    text = str(event).lower()
    assert "password" not in text
    assert "access_token" not in text
    assert "refresh_token" not in text
    assert "authorization" not in text
    assert "bearer" not in text
    assert "secret" not in text


def test_gateway_user_registered_event():
    event = gateway_user_registered("usr-1", "ana@test.com", "Ana")
    assert_common(event, "GatewayUserRegistered")
    assert event["payload"]["status"] == "REGISTERED"
    assert_no_sensitive_data(event)


def test_gateway_user_logged_in_event():
    event = gateway_user_logged_in("usr-1", "ana@test.com")
    assert_common(event, "GatewayUserLoggedIn")
    assert event["payload"]["status"] == "LOGGED_IN"
    assert_no_sensitive_data(event)


def test_gateway_inventory_search_requested_event():
    event = gateway_inventory_search_requested({
        "fecha_inicio": "2026-07-10",
        "fecha_fin": "2026-07-12",
        "ubicacion": "Santiago",
        "precio_max": 120000,
        "capacidad": 2,
    })
    assert_common(event, "GatewayInventorySearchRequested")
    assert event["payload"]["ubicacion"] == "Santiago"
    assert_no_sensitive_data(event)


def test_gateway_reservation_requested_event():
    event = gateway_reservation_requested("usr-1", "hotel-1", "room-1", "2026-07-10", "2026-07-12")
    assert_common(event, "GatewayReservationRequested")
    assert event["payload"]["hotel_id"] == "hotel-1"
    assert_no_sensitive_data(event)
