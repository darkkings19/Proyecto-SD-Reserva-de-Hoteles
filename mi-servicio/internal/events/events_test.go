package events

import (
	"context"
	"testing"
)

func TestNewReservationCreated(t *testing.T) {
	event := NewReservationCreated(sampleReservationData())

	if event.EventID == "" {
		t.Fatal("expected event_id")
	}
	if event.EventType != EventTypeReservationCreated {
		t.Fatalf("unexpected event_type: %s", event.EventType)
	}
	if event.Version != 1 {
		t.Fatalf("unexpected version: %d", event.Version)
	}
	if event.SourceService != SourceReservationService {
		t.Fatalf("unexpected source_service: %s", event.SourceService)
	}
	if event.OccurredAt == "" {
		t.Fatal("expected occurred_at")
	}

	payload, ok := event.Payload.(ReservationCreatedPayload)
	if !ok {
		t.Fatalf("unexpected payload type: %T", event.Payload)
	}
	if payload.Status != "CREATING" {
		t.Fatalf("unexpected status: %s", payload.Status)
	}
	if payload.UserID != "usr-1001" || payload.HotelID != "hotel-001" || payload.RoomTypeID != "room-std-001" {
		t.Fatalf("unexpected reservation payload: %+v", payload)
	}
}

func TestNewReservationConfirmed(t *testing.T) {
	event := NewReservationConfirmed(sampleReservationData())

	if event.EventType != EventTypeReservationConfirmed {
		t.Fatalf("unexpected event_type: %s", event.EventType)
	}

	payload, ok := event.Payload.(ReservationConfirmedPayload)
	if !ok {
		t.Fatalf("unexpected payload type: %T", event.Payload)
	}
	if payload.Status != "CONFIRMED" {
		t.Fatalf("unexpected status: %s", payload.Status)
	}
	if payload.ReservationID != "res-20260705161000" {
		t.Fatalf("unexpected reservation_id: %s", payload.ReservationID)
	}
	if payload.MontoTotal != 150.5 {
		t.Fatalf("unexpected monto_total: %f", payload.MontoTotal)
	}
}

func TestNewReservationFailed(t *testing.T) {
	event := NewReservationFailed(sampleReservationData(), "inventory_lock_failed")

	if event.EventType != EventTypeReservationFailed {
		t.Fatalf("unexpected event_type: %s", event.EventType)
	}

	payload, ok := event.Payload.(ReservationFailedPayload)
	if !ok {
		t.Fatalf("unexpected payload type: %T", event.Payload)
	}
	if payload.Status != "FAILED" {
		t.Fatalf("unexpected status: %s", payload.Status)
	}
	if payload.Reason != "inventory_lock_failed" {
		t.Fatalf("unexpected reason: %s", payload.Reason)
	}
}

func TestNoopPublisherWhenKafkaDisabled(t *testing.T) {
	publisher := NewPublisher(Config{Enabled: false})
	defer publisher.Close()

	err := publisher.Publish(context.Background(), "usr-1001", NewReservationCreated(sampleReservationData()))
	if err != nil {
		t.Fatalf("expected disabled kafka publisher to be no-op, got error: %v", err)
	}
}

func sampleReservationData() ReservationData {
	return ReservationData{
		ReservationID: "res-20260705161000",
		UserID:        "usr-1001",
		HotelID:       "hotel-001",
		RoomTypeID:    "room-std-001",
		FechaInicio:   "2026-07-10",
		FechaFin:      "2026-07-12",
		MontoTotal:    150.5,
	}
}
