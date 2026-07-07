import logging
import time
from typing import Optional

import grpc

from core.domain import Notification
from core.ports import NotificationRepository, NotificationSender
from observability import (
    notification_send_duration_seconds,
    notifications_errors_total,
    notifications_external_total,
    notifications_saved_total,
)
from proto.notifications_pb2 import SendConfirmationResponse
from proto.notifications_pb2_grpc import NotificationServiceServicer

logger = logging.getLogger(__name__)


class NotificationServicer(NotificationServiceServicer):
    def __init__(
        self,
        repository: NotificationRepository,
        sender: Optional[NotificationSender] = None,
    ):
        self.repository = repository
        self.sender = sender
        notifications_saved_total.labels(tipo="CONFIRMACION").inc(0)
        notifications_external_total.labels(status="success").inc(0)
        notifications_external_total.labels(status="failed").inc(0)
        notifications_errors_total.labels(
            operation="send_confirmation",
            error_type="inactive_context",
        ).inc(0)
        notifications_errors_total.labels(
            operation="send_confirmation",
            error_type="internal",
        ).inc(0)

    def SendConfirmation(self, request, context: grpc.ServicerContext) -> SendConfirmationResponse:
        start = time.perf_counter()
        try:
            if not context.is_active():
                logger.warning("SendConfirmation called but context is not active")
                notifications_errors_total.labels(
                    operation="send_confirmation",
                    error_type="inactive_context",
                ).inc()
                return SendConfirmationResponse(success=False)

            self.process_confirmation(
                user_id=request.user_id,
                reservation_id=request.reservation_id,
                tipo=request.tipo,
                email=request.email,
            )

            return SendConfirmationResponse(success=True)
        except Exception as e:
            logger.error("Error en SendConfirmation: %s", e)
            notifications_errors_total.labels(
                operation="send_confirmation",
                error_type="internal",
            ).inc()
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return SendConfirmationResponse(success=False)
        finally:
            notification_send_duration_seconds.observe(time.perf_counter() - start)

    def process_confirmation(
        self,
        user_id: str,
        reservation_id: str,
        tipo: str = "CONFIRMACION",
        email: str = "",
    ) -> bool:
        notification_type = tipo or "CONFIRMACION"
        if self._is_duplicate(user_id, reservation_id, notification_type):
            logger.info(
                "Duplicado detectado: key=reservation-confirmation:%s",
                reservation_id,
            )
            return False

        notification = Notification(
            user_id=user_id,
            reservation_id=reservation_id,
            tipo=notification_type,
            email=email,
        )
        self.repository.save(notification)
        notifications_saved_total.labels(tipo=notification_type).inc()
        logger.info(
            "Notificacion guardada: user=%s, reservation=%s, tipo=%s",
            user_id,
            reservation_id,
            notification_type,
        )

        if self.sender is not None:
            try:
                self.sender.send(notification)
                notifications_external_total.labels(status="success").inc()
            except Exception as e:
                logger.error("Error sending notification externally: %s", e)
                notifications_external_total.labels(status="failed").inc()

        return True

    def _is_duplicate(self, user_id: str, reservation_id: str, tipo: str) -> bool:
        for notification in self.repository.get_by_user(user_id):
            if notification.reservation_id == reservation_id and notification.tipo == tipo:
                return True
        return False
