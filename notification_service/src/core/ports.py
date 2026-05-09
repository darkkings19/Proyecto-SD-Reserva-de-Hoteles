from abc import ABC, abstractmethod
from typing import List
from core.domain import Notification

class NotificationRepository(ABC):
    @abstractmethod
    def save(self, notification: Notification) -> None:
        pass

    @abstractmethod
    def get_by_user(self, user_id: str) -> List[Notification]:
        pass
