from datetime import datetime
from core.domain import Notification

def test_notification_creation():
    notification = Notification(
        user_id="123",
        reservation_id="abc",
        tipo="CONFIRMACION",
        email="test@example.com"
    )
    assert notification.user_id == "123"
    assert notification.reservation_id == "abc"
    assert notification.tipo == "CONFIRMACION"
    assert notification.created_at is None  # Optional — set by repo on save

def test_notification_with_timestamp():
    now = datetime(2026, 5, 10, 12, 0, 0)
    notification = Notification(
        user_id="123",
        reservation_id="abc",
        tipo="CONFIRMACION",
        email="test@example.com",
        created_at=now,
    )
    assert notification.created_at == now
