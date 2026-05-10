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


class ContactInfo:
    """Value object representing a user's contact information."""
    def __init__(self, email: str) -> None:
        self.email = email


class UserContactResolver(ABC):
    """Port for resolving a user's contact info (email, phone, etc.) from their user_id.

    The Notification Service receives user_id in gRPC requests but doesn't
    know the user's email. This port lets each environment resolve it differently:
    - Dev/testing: read from an env var
    - Production: call the User Service via gRPC
    """
    @abstractmethod
    def resolve(self, user_id: str) -> ContactInfo:
        pass
