import grpc
from concurrent import futures
import logging
import signal
import os
import time

from proto.notifications_pb2_grpc import add_NotificationServiceServicer_to_server
from grpc_interface.server import NotificationServicer
from infrastructure.mock_repository import MockNotificationRepository

def serve():
    # Setup repository and servicer
    repo = MockNotificationRepository()
    servicer = NotificationServicer(repo)

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
        logging.info("Server stopped gracefully")
        
    signal.signal(signal.SIGTERM, handle_sigterm)
    signal.signal(signal.SIGINT, handle_sigterm)
    
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    serve()
