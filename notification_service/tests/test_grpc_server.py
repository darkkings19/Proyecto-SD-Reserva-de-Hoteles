import grpc
import pytest
from unittest.mock import Mock, MagicMock
from grpc_interface.server import NotificationServicer
from proto.notifications_pb2 import SendConfirmationRequest
from mocks.mock_repository import MockNotificationRepository
from core.ports import NotificationSender


class MockSender(NotificationSender):
    """Spy that records sent notifications without actually sending."""
    def __init__(self):
        self.sent = []

    def send(self, notification) -> None:
        self.sent.append(notification)


@pytest.fixture
def repo():
    return MockNotificationRepository()

@pytest.fixture
def sender():
    return MockSender()

@pytest.fixture
def servicer(repo):
    return NotificationServicer(repo)

@pytest.fixture
def servicer_with_sender(repo, sender):
    return NotificationServicer(repo, sender=sender)

def test_send_confirmation_success(servicer, repo):
    request = SendConfirmationRequest(
        user_id="user1",
        reservation_id="res1",
        tipo="CONFIRMACION",
        email="test@example.com"
    )
    
    context = Mock(spec=grpc.ServicerContext)
    context.is_active.return_value = True
    response = servicer.SendConfirmation(request, context)
    
    assert response.success is True
    
    notifications = repo.get_by_user("user1")
    assert len(notifications) == 1
    assert notifications[0].reservation_id == "res1"
    assert notifications[0].tipo == "CONFIRMACION"

def test_send_confirmation_exception_handling(servicer):
    request = SendConfirmationRequest(
        user_id="user1",
        reservation_id="res1",
        tipo="CONFIRMACION",
        email="test@example.com"
    )
    
    # Force an exception when saving
    servicer.repository.save = Mock(side_effect=Exception("DB Error"))
    
    context = Mock(spec=grpc.ServicerContext)
    context.is_active.return_value = True
    response = servicer.SendConfirmation(request, context)
    
    assert response.success is False
    context.set_code.assert_called_once_with(grpc.StatusCode.INTERNAL)
    context.set_details.assert_called_once_with("DB Error")

def test_send_confirmation_context_cancelled(servicer, repo):
    request = SendConfirmationRequest(
        user_id="user1",
        reservation_id="res1",
        tipo="CONFIRMACION",
        email="test@example.com"
    )
    
    context = Mock(spec=grpc.ServicerContext)
    context.is_active.return_value = False
    
    response = servicer.SendConfirmation(request, context)
    
    assert response.success is False
    
    # Ensure database wasn't hit
    assert len(repo.get_by_user("user1")) == 0

def test_send_confirmation_calls_sender(servicer_with_sender, sender):
    request = SendConfirmationRequest(
        user_id="user1",
        reservation_id="res1",
        tipo="CONFIRMACION",
        email="test@example.com"
    )

    context = Mock(spec=grpc.ServicerContext)
    context.is_active.return_value = True
    response = servicer_with_sender.SendConfirmation(request, context)

    assert response.success is True
    assert len(sender.sent) == 1
    assert sender.sent[0].user_id == "user1"
    assert sender.sent[0].reservation_id == "res1"
    assert sender.sent[0].tipo == "CONFIRMACION"

def test_send_confirmation_sender_does_not_block_on_error(servicer_with_sender, sender, repo):
    """Sender failure should not affect the gRPC response (fire-and-forget)."""
    sender.send = MagicMock(side_effect=Exception("Resend API down"))

    request = SendConfirmationRequest(
        user_id="user1",
        reservation_id="res1",
        tipo="CONFIRMACION",
        email="test@example.com"
    )

    context = Mock(spec=grpc.ServicerContext)
    context.is_active.return_value = True
    response = servicer_with_sender.SendConfirmation(request, context)

    # Response should still be success (DB saved, email is best-effort)
    assert response.success is True
    assert len(repo.get_by_user("user1")) == 1
