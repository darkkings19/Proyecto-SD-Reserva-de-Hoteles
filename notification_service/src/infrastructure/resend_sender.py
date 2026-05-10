import logging
from typing import Optional
from core.domain import Notification
from core.ports import NotificationSender

logger = logging.getLogger(__name__)


class ResendNotificationSender(NotificationSender):
    """Sends notifications via Resend (email).

    This adapter knows about Resend, but the domain only knows NotificationSender.
    To swap providers, just write a new adapter.
    """

    def __init__(
        self,
        api_key: str,
        from_email: str = "onboarding@resend.dev",
        to_email: Optional[str] = None,
    ):
        self.from_email = from_email
        self.to_email = to_email or "e.loren01@ufromail.cl"
        import resend
        resend.api_key = api_key

    def send(self, notification: Notification) -> None:
        import resend

        if notification.tipo not in ("CONFIRMACION",):
            logger.debug("Skipping email for tipo=%s", notification.tipo)
            return

        subject = "Reserva Confirmada"
        html = (
            f"<p>Hola,</p>"
            f"<p>Tu reserva <strong>{notification.reservation_id}</strong> "
            f"ha sido confirmada exitosamente.</p>"
            f"<p>¡Gracias por preferirnos!</p>"
        )

        try:
            r = resend.Emails.send({
                "from": self.from_email,
                "to": self.to_email,
                "subject": subject,
                "html": html,
            })
            logger.info("Email sent via Resend: id=%s, to=%s", r.get("id"), self.to_email)
        except Exception as e:
            logger.error("Failed to send email via Resend: %s", e)
