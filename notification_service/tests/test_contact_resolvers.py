from unittest.mock import patch, MagicMock
from core.ports import ContactInfo
from infrastructure.env_contact_resolver import EnvContactResolver


def test_env_resolver_returns_configured_email():
    resolver = EnvContactResolver(default_email="test@example.com")
    contact = resolver.resolve("any-user-id")
    assert isinstance(contact, ContactInfo)
    assert contact.email == "test@example.com"


def test_env_resolver_reads_from_env_var(monkeypatch):
    monkeypatch.setenv("DEFAULT_USER_EMAIL", "env-override@example.com")
    resolver = EnvContactResolver(default_email="fallback@example.com")
    contact = resolver.resolve("u1")
    assert contact.email == "env-override@example.com"


def test_env_resolver_ignores_user_id():
    """In dev mode, user_id is irrelevant — all users get the same email."""
    resolver = EnvContactResolver(default_email="dev@example.com")
    assert resolver.resolve("user-1").email == "dev@example.com"
    assert resolver.resolve("user-999").email == "dev@example.com"
