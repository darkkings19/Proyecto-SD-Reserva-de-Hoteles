from __future__ import annotations

import json
import logging
import threading
import time
from dataclasses import dataclass
from typing import TYPE_CHECKING, Any

try:
    from kafka.errors import CommitFailedError, KafkaError, NoBrokersAvailable
except ImportError:  # pragma: no cover - exercised only when kafka-python is absent
    class KafkaError(Exception):
        pass

    class CommitFailedError(KafkaError):
        pass

    class NoBrokersAvailable(KafkaError):
        pass

from kafka_events.models import (
    build_notification_failed_event,
    build_notification_sent_event,
    idempotency_key,
    parse_reservation_confirmed,
    should_process_event,
)

if TYPE_CHECKING:
    from grpc_interface.server import NotificationServicer


logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class KafkaSettings:
    enabled: bool
    bootstrap_servers: str
    reservations_topic: str
    notifications_topic: str
    consumer_group: str
    client_id: str
    default_email: str


class NotificationKafkaWorker:
    INITIAL_RETRY_SECONDS = 2
    MAX_RETRY_SECONDS = 30

    def __init__(self, settings: KafkaSettings, servicer: "NotificationServicer"):
        self.settings = settings
        self.servicer = servicer
        self._stop_event = threading.Event()
        self._thread: threading.Thread | None = None
        self._consumer: Any | None = None
        self._producer: Any | None = None

    def start(self) -> None:
        if not self.settings.enabled:
            logger.info("[Kafka] Consumer deshabilitado por KAFKA_ENABLED=false")
            return

        self._thread = threading.Thread(target=self._run, name="notification-kafka-worker", daemon=True)
        self._thread.start()
        logger.info("[Kafka] Consumer iniciado")

    def stop(self) -> None:
        self._stop_event.set()
        self._close_clients()
        if self._thread is not None:
            self._thread.join(timeout=5)

    def _run(self) -> None:
        retry_seconds = self.INITIAL_RETRY_SECONDS

        while not self._stop_event.is_set():
            try:
                self._connect_clients()
                retry_seconds = self.INITIAL_RETRY_SECONDS
                logger.info(
                    "[Kafka] Consumer conectado: topic=%s group=%s",
                    self.settings.reservations_topic,
                    self.settings.consumer_group,
                )
                self._consume_until_stopped()
            except NoBrokersAvailable as exc:
                if self._stop_event.is_set():
                    logger.info("[Kafka] Worker detenido durante cierre")
                else:
                    self._log_retry("Kafka no disponible", exc, retry_seconds)
            except KafkaError as exc:
                if self._stop_event.is_set():
                    logger.info("[Kafka] Worker detenido durante cierre")
                else:
                    self._log_retry("Error de Kafka", exc, retry_seconds)
            except Exception as exc:
                if self._stop_event.is_set():
                    logger.info("[Kafka] Worker detenido durante cierre")
                else:
                    self._log_retry("Error inesperado en worker Kafka", exc, retry_seconds)
            finally:
                self._close_clients()

            if not self._stop_event.is_set():
                if self._stop_event.wait(retry_seconds):
                    break
                retry_seconds = min(retry_seconds * 2, self.MAX_RETRY_SECONDS)

        logger.info("[Kafka] Worker detenido")

    def _connect_clients(self) -> None:
        try:
            from kafka import KafkaConsumer, KafkaProducer
        except ImportError as exc:
            raise KafkaError(f"kafka-python no esta instalado: {exc}") from exc

        self._consumer = KafkaConsumer(
            self.settings.reservations_topic,
            bootstrap_servers=self._servers(),
            group_id=self.settings.consumer_group,
            client_id=self.settings.client_id,
            enable_auto_commit=False,
            auto_offset_reset="earliest",
            consumer_timeout_ms=1000,
            value_deserializer=lambda raw: json.loads(raw.decode("utf-8")),
        )
        self._producer = KafkaProducer(
            bootstrap_servers=self._servers(),
            client_id=self.settings.client_id,
            value_serializer=lambda value: json.dumps(value).encode("utf-8"),
            key_serializer=lambda value: value.encode("utf-8"),
            request_timeout_ms=5000,
            retries=3,
        )

    def _consume_until_stopped(self) -> None:
        if self._consumer is None:
            return

        while not self._stop_event.is_set():
            for message in self._consumer:
                if self._stop_event.is_set():
                    break

                self.handle_event(message.value)
                self._commit_message()

            time.sleep(0.05)

    def _commit_message(self) -> None:
        if self._consumer is None:
            return

        try:
            self._consumer.commit()
        except CommitFailedError as exc:
            logger.warning("[Kafka] Commit omitido por rebalance/shutdown: %s", exc)
        except KafkaError as exc:
            logger.warning("[Kafka] Error al confirmar offset: %s", exc)

    def _close_clients(self) -> None:
        consumer = self._consumer
        producer = self._producer
        self._consumer = None
        self._producer = None

        if consumer is not None:
            try:
                consumer.close(autocommit=False)
            except CommitFailedError as exc:
                logger.warning("[Kafka] Commit omitido durante cierre: %s", exc)
            except Exception as exc:
                logger.warning("[Kafka] Error cerrando consumer: %s", exc)

        if producer is not None:
            try:
                producer.close(timeout=5)
            except Exception as exc:
                logger.warning("[Kafka] Error cerrando producer: %s", exc)

    def _log_retry(self, message: str, exc: Exception, retry_seconds: int) -> None:
        logger.warning(
            "[Kafka] %s: %s. Reintentando en %ss",
            message,
            exc,
            retry_seconds,
        )

    def handle_event(self, event: dict[str, Any]) -> str:
        event_type = event.get("event_type")
        logger.info("[Kafka] Evento recibido: %s", event_type)

        if not should_process_event(event):
            logger.info("[Kafka] Evento ignorado: %s", event_type)
            return "ignored"

        reservation = None
        try:
            reservation = parse_reservation_confirmed(
                event,
                default_email=self.settings.default_email,
            )
            key = idempotency_key(reservation.reservation_id)

            processed = self.servicer.process_confirmation(
                user_id=reservation.user_id,
                reservation_id=reservation.reservation_id,
                tipo="CONFIRMACION",
                email=reservation.email,
            )
            if not processed:
                logger.info("[Kafka] Duplicado detectado: %s", key)
                return "duplicate"

            logger.info("[Kafka] ReservationConfirmed procesado: %s", reservation.reservation_id)
            self._publish_notification_event(
                key=key,
                event=build_notification_sent_event(reservation, correlation_id=event.get("event_id")),
            )
            logger.info("[Kafka] NotificationSent publicado: %s", reservation.reservation_id)
            return "processed"
        except Exception as exc:
            logger.error("[Kafka] Error procesando ReservationConfirmed: %s", exc)
            failed_event = build_notification_failed_event(
                reservation,
                reason=str(exc),
                correlation_id=event.get("event_id"),
            )
            self._publish_notification_event(
                key=idempotency_key(reservation.reservation_id) if reservation else str(event.get("event_id") or "unknown"),
                event=failed_event,
            )
            return "failed"

    def _publish_notification_event(self, key: str, event: dict[str, Any]) -> None:
        if self._producer is None:
            logger.warning("[Kafka] Producer no inicializado; no se publico %s", event.get("event_type"))
            return
        try:
            self._producer.send(self.settings.notifications_topic, key=key, value=event)
            self._producer.flush(timeout=2)
        except Exception as exc:
            logger.error("[Kafka] Error de Kafka publicando %s: %s", event.get("event_type"), exc)

    def _servers(self) -> list[str]:
        return [server.strip() for server in self.settings.bootstrap_servers.split(",") if server.strip()]
