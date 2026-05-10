from unittest.mock import patch, MagicMock
from core.domain import Notification
from core.ports import UserContactResolver, ContactInfo
from infrastructure.resend_sender import ResendNotificationSender


class FakeResolver(UserContactResolver):
    """Always resolves to a fixed email — useful for tests."""
    def __init__(self, email: str = "user@example.com"):
        self._email = email

    def resolve(self, user_id: str) -> ContactInfo:
        return ContactInfo(email=self._email)


def test_resend_sender_sends_on_confirmacion():
    sender = ResendNotificationSender(
        api_key="test_key",
        from_email="test@example.com",
        contact_resolver=FakeResolver("user@example.com"),
    )

    notification = Notification(
        user_id="u1",
        reservation_id="res-1",
        tipo="CONFIRMACION",
    )

    with patch("resend.Emails.send", return_value={"id": "email_123"}) as mock_send:
        sender.send(notification)

    mock_send.assert_called_once()
    call_kwargs = mock_send.call_args[0][0]
    assert call_kwargs["to"] == "user@example.com"
    assert call_kwargs["from"] == "test@example.com"
    assert call_kwargs["subject"] == "Reserva Confirmada"
    assert "res-1" in call_kwargs["html"]


def test_resend_sender_skips_unknown_tipo():
    sender = ResendNotificationSender(
        api_key="test_key",
        contact_resolver=FakeResolver(),
    )

    notification = Notification(
        user_id="u1",
        reservation_id="res-1",
        tipo="CANCELACION",
    )

    with patch("resend.Emails.send") as mock_send:
        sender.send(notification)

    mock_send.assert_not_called()


def test_resend_sender_does_not_raise_on_api_error():
    sender = ResendNotificationSender(
        api_key="test_key",
        contact_resolver=FakeResolver(),
    )

    notification = Notification(
        user_id="u1",
        reservation_id="res-1",
        tipo="CONFIRMACION",
    )

    with patch("resend.Emails.send", side_effect=Exception("API timeout")):
        # Should not raise — errors are logged, not propagated
        sender.send(notification)


def test_resend_sender_logs_resolver_failure():
    """If the resolver fails, the sender logs it and skips — does not crash."""
    class BrokenResolver(UserContactResolver):
        def resolve(self, user_id: str) -> ContactInfo:
            raise RuntimeError("User service unavailable")

    sender = ResendNotificationSender(
        api_key="test_key",
        contact_resolver=BrokenResolver(),
    )

    notification = Notification(
        user_id="u1",
        reservation_id="res-1",
        tipo="CONFIRMACION",
    )

    with patch("resend.Emails.send") as mock_send:
        sender.send(notification)

    mock_send.assert_not_called()
