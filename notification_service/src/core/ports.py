from abc import ABC, abstractmethod
from typing import List, Optional
from core.domain import Notification


class NotificationRepository(ABC):
    """Port for persisting notifications."""
    @abstractmethod
    def save(self, notification: Notification) -> None:
        pass

    @abstractmethod
    def get_by_user(self, user_id: str) -> List[Notification]:
        pass


class NotificationSender(ABC):
    """Port for sending notifications via external channels (email, SMS, etc.)."""
    @abstractmethod
    def send(self, notification: Notification) -> None:
        pass
