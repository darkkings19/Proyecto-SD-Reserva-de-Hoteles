import logging
from typing import Optional
from core.domain import Notification
from core.ports import NotificationSender, UserContactResolver

logger = logging.getLogger(__name__)


class ResendNotificationSender(NotificationSender):
    """Sends notifications via Resend (email).

    Uses a UserContactResolver to determine the recipient's email — the sender
    doesn't need to know how the email is obtained. Swap resolvers per environment.
    """

    def __init__(
        self,
        api_key: str,
        contact_resolver: UserContactResolver,
        from_email: str = "onboarding@resend.dev",
    ):
        import resend

        resend.api_key = api_key
        self.from_email = from_email
        self.contact_resolver = contact_resolver

    def send(self, notification: Notification) -> None:
        import resend

        if notification.tipo not in ("CONFIRMACION",):
            logger.debug("Skipping email for tipo=%s", notification.tipo)
            return

        # Resolve the recipient's email
        try:
            contact = self.contact_resolver.resolve(notification.user_id)
        except Exception as e:
            logger.error(
                "Cannot send email for user %s: failed to resolve contact: %s",
                notification.user_id,
                e,
            )
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
                "to": contact.email,
                "subject": subject,
                "html": html,
            })
            logger.info(
                "Email sent via Resend: id=%s, to=%s",
                r.get("id"),
                contact.email,
            )
        except Exception as e:
            logger.error("Failed to send email via Resend: %s", e)
