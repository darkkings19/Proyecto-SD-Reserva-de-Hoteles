import grpc
import pytest
from unittest.mock import Mock
from grpc_interface.server import NotificationServicer
from proto.notifications_pb2 import SendConfirmationRequest
from mocks.mock_repository import MockNotificationRepository


@pytest.fixture
def repo():
    return MockNotificationRepository()

@pytest.fixture
def servicer(repo):
    return NotificationServicer(repo)

def test_send_confirmation_success(servicer, repo):
    request = SendConfirmationRequest(
        user_id="user1",
        reservation_id="res1",
        tipo="CONFIRMACION"
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
        tipo="CONFIRMACION"
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
        tipo="CONFIRMACION"
    )
    
    context = Mock(spec=grpc.ServicerContext)
    context.is_active.return_value = False
    
    response = servicer.SendConfirmation(request, context)
    
    assert response.success is False
    
    # Ensure database wasn't hit
    assert len(repo.get_by_user("user1")) == 0
