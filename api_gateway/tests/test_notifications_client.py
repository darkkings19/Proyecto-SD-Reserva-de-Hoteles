import pytest
from unittest.mock import AsyncMock, patch
import grpc
from grpc_clients.notifications_client import NotificationClient

@pytest.mark.asyncio
async def test_send_confirmation_success():
    with patch('grpc_clients.notifications_client.grpc.aio.insecure_channel'):
        mock_stub = AsyncMock()
        mock_response = AsyncMock()
        mock_response.success = True
        mock_stub.SendConfirmation.return_value = mock_response
        
        with patch('grpc_clients.notifications_client.notifications_pb2_grpc.NotificationServiceStub', return_value=mock_stub):
            client = NotificationClient("fake-host:50051")
            success = await client.send_confirmation("user1", "res1", "CONFIRMACION")
            
            assert success is True
            mock_stub.SendConfirmation.assert_called_once()

@pytest.mark.asyncio
async def test_send_confirmation_failure():
    with patch('grpc_clients.notifications_client.grpc.aio.insecure_channel'):
        mock_stub = AsyncMock()
        mock_stub.SendConfirmation.side_effect = grpc.aio.AioRpcError(
            code=grpc.StatusCode.UNAVAILABLE,
            initial_metadata=None,
            trailing_metadata=None,
            details="Service unavailable"
        )
        
        with patch('grpc_clients.notifications_client.notifications_pb2_grpc.NotificationServiceStub', return_value=mock_stub):
            client = NotificationClient("fake-host:50051")
            success = await client.send_confirmation("user1", "res1", "CONFIRMACION")
            
            assert success is False
