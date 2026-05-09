from core.domain import Notification

def test_notification_creation():
    notification = Notification(
        user_id="123",
        reservation_id="abc",
        tipo="CONFIRMACION"
    )
    assert notification.user_id == "123"
    assert notification.reservation_id == "abc"
    assert notification.tipo == "CONFIRMACION"
