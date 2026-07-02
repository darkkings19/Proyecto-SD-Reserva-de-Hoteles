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

            notification = Notification(
                user_id=request.user_id,
                reservation_id=request.reservation_id,
                tipo=request.tipo,
                email=request.email,
            )
            self.repository.save(notification)
            notifications_saved_total.labels(tipo=request.tipo or "unknown").inc()
            logger.info(
                "Notificacion guardada: user=%s, reservation=%s, tipo=%s",
                request.user_id,
                request.reservation_id,
                request.tipo,
            )

            if self.sender is not None:
                try:
                    self.sender.send(notification)
                    notifications_external_total.labels(status="success").inc()
                except Exception as e:
                    logger.error("Error sending notification externally: %s", e)
                    notifications_external_total.labels(status="failed").inc()

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
