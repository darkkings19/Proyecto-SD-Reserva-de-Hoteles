from unittest.mock import patch, MagicMock
from core.ports import ContactInfo
from infrastructure.grpc_contact_resolver import GrpcUserContactResolver
from proto.user_service_pb2 import UserResponse, User


def _fake_user_response(email: str) -> UserResponse:
    return UserResponse(user=User(id="u1", nombre="Test", email=email))


def test_grpc_resolver_returns_email_from_user_service():
    mock_stub = MagicMock()
    mock_stub.GetUser.return_value = _fake_user_response("real@example.com")

    resolver = GrpcUserContactResolver(user_service_host="localhost:9090")
    resolver._stub = mock_stub  # inject mock

    contact = resolver.resolve("u1")
    assert isinstance(contact, ContactInfo)
    assert contact.email == "real@example.com"
    mock_stub.GetUser.assert_called_once()


def test_grpc_resolver_raises_on_service_error():
    mock_stub = MagicMock()
    import grpc
    mock_stub.GetUser.side_effect = grpc.RpcError()

    resolver = GrpcUserContactResolver(user_service_host="localhost:9090")
    resolver._stub = mock_stub

    import pytest
    with pytest.raises(grpc.RpcError):
        resolver.resolve("unknown-user")
