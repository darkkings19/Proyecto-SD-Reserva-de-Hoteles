from core.domain import Notification
from mocks.mock_repository import MockNotificationRepository

def test_save_and_retrieve_notification():
    repo = MockNotificationRepository()
    notification = Notification(user_id="1", reservation_id="100", tipo="CONFIRMACION", email="test@example.com")
    repo.save(notification)

    notifications = repo.get_by_user("1")
    assert len(notifications) == 1
    assert notifications[0].reservation_id == "100"
    assert notifications[0].created_at is not None  # Auto-assigned by mock

def test_retrieve_non_existent_user():
    repo = MockNotificationRepository()
    notifications = repo.get_by_user("999")
    assert notifications == []
