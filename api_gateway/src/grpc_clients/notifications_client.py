import grpc.aio
import logging
from proto import notifications_pb2
from proto import notifications_pb2_grpc

class NotificationClient:
    def __init__(self, host: str):
        self.host = host

    async def send_confirmation(self, user_id: str, reservation_id: str, tipo: str) -> bool:
        try:
            async with grpc.aio.insecure_channel(self.host) as channel:
                stub = notifications_pb2_grpc.NotificationServiceStub(channel)
                request = notifications_pb2.SendConfirmationRequest(
                    user_id=user_id,
                    reservation_id=reservation_id,
                    tipo=tipo
                )
                response = await stub.SendConfirmation(request, timeout=2.0)
                return response.success
        except grpc.aio.AioRpcError as e:
            logging.error(f"gRPC call failed: {e.code()} - {e.details()}")
            return False
        except Exception as e:
            logging.error(f"Unexpected error calling Notification Service: {e}")
            return False