import os
import logging
from core.ports import UserContactResolver, ContactInfo

logger = logging.getLogger(__name__)


class EnvContactResolver(UserContactResolver):
    """Resolves contact info from environment variables.

    For dev/testing: returns a hardcoded email from DEFAULT_USER_EMAIL.
    Falls back to a default if the env var is not set.
    """

    def __init__(self, default_email: str = "e.loren01@ufromail.cl"):
        self.default_email = os.environ.get("DEFAULT_USER_EMAIL", default_email)
        logger.info(
            "EnvContactResolver initialized — all users will receive email at %s",
            self.default_email,
        )

    def resolve(self, user_id: str) -> ContactInfo:
        # In dev mode we ignore user_id and return the configured email
        _ = user_id  # explicitly unused
        return ContactInfo(email=self.default_email)
