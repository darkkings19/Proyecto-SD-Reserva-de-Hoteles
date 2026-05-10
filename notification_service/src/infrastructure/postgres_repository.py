from typing import List
from core.domain import Notification
from core.ports import NotificationRepository
from psycopg_pool import ConnectionPool

class PostgresNotificationRepository(NotificationRepository):
    def __init__(self, pool: ConnectionPool):
        self.pool = pool

    def initialize_schema(self) -> None:
        create_table_sql = """
        CREATE TABLE IF NOT EXISTS notifications (
            id SERIAL PRIMARY KEY,
            user_id VARCHAR(255) NOT NULL,
            reservation_id VARCHAR(255) NOT NULL,
            tipo VARCHAR(50) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE (reservation_id, tipo)
        );
        """
        with self.pool.connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute(create_table_sql)
            conn.commit()

    def save(self, notification: Notification) -> None:
        sql = """
        INSERT INTO notifications (user_id, reservation_id, tipo)
        VALUES (%s, %s, %s)
        ON CONFLICT (reservation_id, tipo) DO NOTHING;
        """
        with self.pool.connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute(
                    sql,
                    (notification.user_id, notification.reservation_id, notification.tipo)
                )
            conn.commit()

    def get_by_user(self, user_id: str) -> List[Notification]:
        sql = """
        SELECT user_id, reservation_id, tipo, created_at
        FROM notifications
        WHERE user_id = %s
        ORDER BY created_at DESC;
        """
        notifications = []
        with self.pool.connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute(sql, (user_id,))
                rows = cursor.fetchall()
                for row in rows:
                    notifications.append(
                        Notification(
                            user_id=row[0],
                            reservation_id=row[1],
                            tipo=row[2],
                            created_at=row[3],
                        )
                    )
        return notifications
