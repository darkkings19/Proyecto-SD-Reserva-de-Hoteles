import json
import logging
import os
import threading
from datetime import datetime, timezone
from typing import Any
from uuid import uuid4

try:
    from kafka import KafkaProducer
except ImportError:  # pragma: no cover
    KafkaProducer = None


SOURCE_SERVICE = "api-gateway"


def build_event(event_type: str, payload: dict[str, Any], correlation_id: str | None = None) -> dict[str, Any]:
    event = {
        "event_id": str(uuid4()),
        "event_type": event_type,
        "version": 1,
        "source_service": SOURCE_SERVICE,
        "occurred_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "payload": payload,
    }
    if correlation_id:
        event["correlation_id"] = correlation_id
    return event


def gateway_user_registered(user_id: str, email: str, nombre: str) -> dict[str, Any]:
    return build_event("GatewayUserRegistered", {
        "user_id": user_id,
        "email": email,
        "nombre": nombre,
        "status": "REGISTERED",
    })


def gateway_user_logged_in(user_id: str, email: str) -> dict[str, Any]:
    return build_event("GatewayUserLoggedIn", {
        "user_id": user_id,
        "email": email,
        "status": "LOGGED_IN",
    })


def gateway_inventory_search_requested(filters: dict[str, Any]) -> dict[str, Any]:
    return build_event("GatewayInventorySearchRequested", {
        "fecha_inicio": filters.get("fecha_inicio", ""),
        "fecha_fin": filters.get("fecha_fin", ""),
        "ubicacion": filters.get("ubicacion", ""),
        "precio_max": filters.get("precio_max", 0),
        "capacidad": filters.get("capacidad", 0),
        "status": "REQUESTED",
    })


def gateway_reservation_requested(
    user_id: str,
    hotel_id: str,
    room_type_id: str,
    fecha_inicio: str,
    fecha_fin: str,
) -> dict[str, Any]:
    return build_event("GatewayReservationRequested", {
        "user_id": user_id,
        "hotel_id": hotel_id,
        "room_type_id": room_type_id,
        "fecha_inicio": fecha_inicio,
        "fecha_fin": fecha_fin,
        "status": "REQUESTED",
    })


class GatewayKafkaPublisher:
    def __init__(self) -> None:
        self.enabled = os.getenv("KAFKA_ENABLED", "false").lower() in {"true", "1", "yes", "y", "on"}
        self.topic = os.getenv("KAFKA_GATEWAY_TOPIC", "origenx.gateway.events")
        self.bootstrap_servers = os.getenv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")
        self.client_id = os.getenv("KAFKA_CLIENT_ID", SOURCE_SERVICE)
        self._producer = None
        self._lock = threading.Lock()

    def publish(self, key: str, event: dict[str, Any]) -> None:
        if not self.enabled:
            return
        try:
            producer = self._get_producer()
            producer.send(self.topic, key=key, value=event)
            producer.flush(timeout=2)
            logging.info("[Gateway][Kafka] Evento publicado: %s key=%s", event.get("event_type"), key)
        except Exception as exc:
            logging.warning("[Gateway][Kafka] Error publicando %s: %s", event.get("event_type"), exc)

    def _get_producer(self):
        if KafkaProducer is None:
            raise RuntimeError("kafka-python no esta instalado")
        if self._producer is None:
            with self._lock:
                if self._producer is None:
                    self._producer = KafkaProducer(
                        bootstrap_servers=self.bootstrap_servers,
                        client_id=self.client_id,
                        value_serializer=lambda value: json.dumps(value).encode("utf-8"),
                        key_serializer=lambda value: value.encode("utf-8"),
                        request_timeout_ms=5000,
                        retries=3,
                    )
        return self._producer

    def close(self) -> None:
        if self._producer is not None:
            self._producer.close(timeout=5)


publisher = GatewayKafkaPublisher()
