package events

import (
	"context"
	"testing"
)

func TestNewInventoryStockBlocked(t *testing.T) {
	event := NewInventoryStockBlocked(sampleStockData(OperationBlock))

	assertCommonEvent(t, event, EventTypeInventoryStockBlocked)

	payload, ok := event.Payload.(InventoryStockPayload)
	if !ok {
		t.Fatalf("unexpected payload type: %T", event.Payload)
	}
	if payload.HotelID != "hotel-001" || payload.RoomTypeID != "room-std-001" {
		t.Fatalf("unexpected inventory payload: %+v", payload)
	}
	if payload.Quantity != 1 || payload.Operation != OperationBlock || payload.Status != StatusBlocked {
		t.Fatalf("unexpected blocked payload: %+v", payload)
	}
}

func TestNewInventoryStockReleased(t *testing.T) {
	event := NewInventoryStockReleased(sampleStockData(OperationRelease))

	assertCommonEvent(t, event, EventTypeInventoryStockReleased)

	payload, ok := event.Payload.(InventoryStockPayload)
	if !ok {
		t.Fatalf("unexpected payload type: %T", event.Payload)
	}
	if payload.Quantity != 1 || payload.Operation != OperationRelease || payload.Status != StatusReleased {
		t.Fatalf("unexpected released payload: %+v", payload)
	}
}

func TestNewInventoryStockFailed(t *testing.T) {
	event := NewInventoryStockFailed(sampleStockData(OperationBlock), "no hay stock suficiente")

	assertCommonEvent(t, event, EventTypeInventoryStockFailed)

	payload, ok := event.Payload.(InventoryStockFailedPayload)
	if !ok {
		t.Fatalf("unexpected payload type: %T", event.Payload)
	}
	if payload.Status != StatusFailed {
		t.Fatalf("unexpected status: %s", payload.Status)
	}
	if payload.Operation != OperationBlock {
		t.Fatalf("unexpected operation: %s", payload.Operation)
	}
	if payload.Reason != "no hay stock suficiente" {
		t.Fatalf("unexpected reason: %s", payload.Reason)
	}
}

func TestNoopPublisherWhenKafkaDisabled(t *testing.T) {
	publisher := NewPublisher(Config{Enabled: false})
	defer publisher.Close()

	err := publisher.Publish(context.Background(), "room-std-001", NewInventoryStockBlocked(sampleStockData(OperationBlock)))
	if err != nil {
		t.Fatalf("expected disabled kafka publisher to be no-op, got error: %v", err)
	}
}

func assertCommonEvent(t *testing.T, event Event, expectedType string) {
	t.Helper()

	if event.EventID == "" {
		t.Fatal("expected event_id")
	}
	if event.EventType != expectedType {
		t.Fatalf("unexpected event_type: %s", event.EventType)
	}
	if event.Version != 1 {
		t.Fatalf("unexpected version: %d", event.Version)
	}
	if event.SourceService != SourceInventoryService {
		t.Fatalf("unexpected source_service: %s", event.SourceService)
	}
	if event.OccurredAt == "" {
		t.Fatal("expected occurred_at")
	}
}

func sampleStockData(operation string) StockData {
	return StockData{
		HotelID:    "hotel-001",
		RoomTypeID: "room-std-001",
		Quantity:   1,
		Operation:  operation,
	}
}
