import logging
import grpc
from typing import Optional
from proto.notifications_pb2 import SendConfirmationResponse
from proto.notifications_pb2_grpc import NotificationServiceServicer
from core.ports import NotificationRepository, NotificationSender
from core.domain import Notification

logger = logging.getLogger(__name__)

class NotificationServicer(NotificationServiceServicer):
    def __init__(
        self,
        repository: NotificationRepository,
        sender: Optional[NotificationSender] = None,
    ):
        self.repository = repository
        self.sender = sender

    def SendConfirmation(self, request, context: grpc.ServicerContext) -> SendConfirmationResponse:
        try:
            if not context.is_active():
                logger.warning("SendConfirmation called but context is not active")
                return SendConfirmationResponse(success=False)

            notification = Notification(
                user_id=request.user_id,
                reservation_id=request.reservation_id,
                tipo=request.tipo,
            )
            self.repository.save(notification)
            logger.info(
                "Notificación guardada: user=%s, reservation=%s, tipo=%s",
                request.user_id, request.reservation_id, request.tipo,
            )

            # Fire-and-forget: send via external channel if configured
            if self.sender is not None:
                try:
                    self.sender.send(notification)
                except Exception as e:
                    logger.error("Error sending notification externally: %s", e)

            return SendConfirmationResponse(success=True)
        except Exception as e:
            logger.error("Error en SendConfirmation: %s", e)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return SendConfirmationResponse(success=False)
