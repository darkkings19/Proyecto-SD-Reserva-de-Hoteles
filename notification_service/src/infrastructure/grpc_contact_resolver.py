import logging
import grpc
from core.ports import UserContactResolver, ContactInfo
from proto.user_service_pb2 import GetUserRequest
from proto.user_service_pb2_grpc import UserServiceStub

logger = logging.getLogger(__name__)


class GrpcUserContactResolver(UserContactResolver):
    """Resolves a user's email by calling the User Service via gRPC.

    This is the production adapter — it goes to the source of truth.
    """

    def __init__(self, user_service_host: str = "user-service:9090"):
        self._host = user_service_host
        self._channel = grpc.insecure_channel(self._host)
        self._stub = UserServiceStub(self._channel)
        logger.info("GrpcUserContactResolver initialized — target: %s", self._host)

    def resolve(self, user_id: str) -> ContactInfo:
        try:
            response = self._stub.GetUser(GetUserRequest(id=user_id))
            email = response.user.email
            logger.debug("Resolved email for user %s: %s", user_id, email)
            return ContactInfo(email=email)
        except grpc.RpcError as e:
            logger.error("Failed to resolve contact for user %s: %s", user_id, e)
            raise
