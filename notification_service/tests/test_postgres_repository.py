import pytest
from unittest.mock import MagicMock
from core.domain import Notification
from infrastructure.postgres_repository import PostgresNotificationRepository

@pytest.fixture
def mock_pool():
    pool = MagicMock()
    conn = MagicMock()
    cursor = MagicMock()
    
    # Setup context managers
    pool.connection.return_value.__enter__.return_value = conn
    conn.cursor.return_value.__enter__.return_value = cursor
    
    # Store cursor for assertions
    pool._mock_cursor = cursor
    return pool

def test_save_notification(mock_pool):
    repo = PostgresNotificationRepository(mock_pool)
    notification = Notification(user_id="user1", reservation_id="res1", tipo="CONFIRMACION")
    
    repo.save(notification)
    
    mock_pool._mock_cursor.execute.assert_called_once()
    sql, params = mock_pool._mock_cursor.execute.call_args[0]
    
    assert "INSERT INTO notifications" in sql
    assert "ON CONFLICT (reservation_id, tipo) DO NOTHING" in sql
    assert params == ("user1", "res1", "CONFIRMACION")

def test_get_by_user(mock_pool):
    repo = PostgresNotificationRepository(mock_pool)
    
    # Mock the fetchall return value
    mock_pool._mock_cursor.fetchall.return_value = [
        ("user1", "res1", "CONFIRMACION")
    ]
    
    notifications = repo.get_by_user("user1")
    
    mock_pool._mock_cursor.execute.assert_called_once()
    sql, params = mock_pool._mock_cursor.execute.call_args[0]
    
    assert "SELECT user_id, reservation_id, tipo" in sql
    assert "FROM notifications" in sql
    assert "WHERE user_id = %s" in sql
    assert params == ("user1",)
    
    assert len(notifications) == 1
    assert notifications[0].user_id == "user1"
    assert notifications[0].reservation_id == "res1"
    assert notifications[0].tipo == "CONFIRMACION"

def test_initialize_schema(mock_pool):
    repo = PostgresNotificationRepository(mock_pool)
    repo.initialize_schema()
    
    mock_pool._mock_cursor.execute.assert_called_once()
    sql = mock_pool._mock_cursor.execute.call_args[0][0]
    assert "CREATE TABLE IF NOT EXISTS notifications" in sql
    assert "UNIQUE (reservation_id, tipo)" in sql
