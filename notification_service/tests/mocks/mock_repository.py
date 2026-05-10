from datetime import datetime, timezone
from typing import List, Dict
from core.domain import Notification
from core.ports import NotificationRepository

class MockNotificationRepository(NotificationRepository):
    def __init__(self) -> None:
        self._storage: Dict[str, List[Notification]] = {}

    def save(self, notification: Notification) -> None:
        if notification.created_at is None:
            notification.created_at = datetime.now(timezone.utc)
        if notification.user_id not in self._storage:
            self._storage[notification.user_id] = []
        self._storage[notification.user_id].append(notification)

    def get_by_user(self, user_id: str) -> List[Notification]:
        return self._storage.get(user_id, [])
