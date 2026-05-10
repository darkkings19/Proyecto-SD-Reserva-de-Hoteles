import grpc
from concurrent import futures
import logging
import signal
import os
import time
from dotenv import load_dotenv

from proto.notifications_pb2_grpc import add_NotificationServiceServicer_to_server
from grpc_interface.server import NotificationServicer
from infrastructure.postgres_repository import PostgresNotificationRepository
from infrastructure.resend_sender import ResendNotificationSender
from psycopg_pool import ConnectionPool

def serve():
    load_dotenv()
    
    db_host = os.environ.get("DB_HOST", "localhost")
    db_port = os.environ.get("DB_PORT", "5432")
    db_user = os.environ.get("DB_USER", "postgres")
    db_password = os.environ.get("DB_PASSWORD", "postgres")
    db_name = os.environ.get("DB_NAME", "notificaciones_db")
    
    conninfo = f"dbname={db_name} user={db_user} password={db_password} host={db_host} port={db_port}"

    # Setup repository and servicer
    try:
        pool = ConnectionPool(conninfo=conninfo, min_size=1, max_size=10)
    except Exception as e:
        logging.error(f"Failed to initialize database connection pool: {e}")
        # In a real scenario we might exit, but to allow CI/tests without DB to pass or retry, we log it.
        # We will allow it to fail at runtime if DB is strictly required.
        raise e

    repo = PostgresNotificationRepository(pool)
    
    # Ensure schema exists (Simple migration approach for this delivery)
    try:
        repo.initialize_schema()
        logging.info("Database schema verified")
    except Exception as e:
        logging.error(f"Failed to initialize schema: {e}")

    # ─ Notification Sender (opcional — external channel como email) ──
    sender = None
    resend_api_key = os.environ.get("RESEND_API_KEY")
    if resend_api_key:
        sender = ResendNotificationSender(
            api_key=resend_api_key,
            from_email=os.environ.get("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
            to_email=os.environ.get("RESEND_TO_EMAIL"),
        )
        logging.info("NotificationSender enabled via Resend")

    servicer = NotificationServicer(repo, sender=sender)

    # Setup server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    add_NotificationServiceServicer_to_server(servicer, server)
    
    port = os.environ.get("PORT", "50051")
    server.add_insecure_port(f"[::]:{port}")
    server.start()
    logging.info(f"Notification Service started on port {port}")
    
    # Graceful shutdown
    def handle_sigterm(*_):
        logging.info("Received shutdown signal")
        all_rpcs_done_event = server.stop(30)
        all_rpcs_done_event.wait(30)
        pool.close()
        logging.info("Server stopped gracefully")
        
    signal.signal(signal.SIGTERM, handle_sigterm)
    signal.signal(signal.SIGINT, handle_sigterm)
    
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)
        pool.close()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    serve()
