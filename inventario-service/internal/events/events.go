package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeInventoryStockBlocked  = "InventoryStockBlocked"
	EventTypeInventoryStockReleased = "InventoryStockReleased"
	EventTypeInventoryStockFailed   = "InventoryStockFailed"

	OperationBlock   = "BLOCK"
	OperationRelease = "RELEASE"

	StatusBlocked  = "BLOCKED"
	StatusReleased = "RELEASED"
	StatusFailed   = "FAILED"

	SourceInventoryService = "inventario-service"
)

type Event struct {
	EventID       string      `json:"event_id"`
	EventType     string      `json:"event_type"`
	Version       int         `json:"version"`
	SourceService string      `json:"source_service"`
	OccurredAt    string      `json:"occurred_at"`
	CorrelationID string      `json:"correlation_id,omitempty"`
	Payload       interface{} `json:"payload"`
}

type StockData struct {
	HotelID    string
	RoomTypeID string
	Quantity   int
	Operation  string
}

type InventoryStockPayload struct {
	HotelID    string `json:"hotel_id"`
	RoomTypeID string `json:"room_type_id"`
	Quantity   int    `json:"quantity"`
	Operation  string `json:"operation"`
	Status     string `json:"status"`
}

type InventoryStockFailedPayload struct {
	HotelID    string `json:"hotel_id"`
	RoomTypeID string `json:"room_type_id"`
	Quantity   int    `json:"quantity"`
	Operation  string `json:"operation"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}

func NewInventoryStockBlocked(data StockData) Event {
	return newEvent(EventTypeInventoryStockBlocked, InventoryStockPayload{
		HotelID:    data.HotelID,
		RoomTypeID: data.RoomTypeID,
		Quantity:   data.Quantity,
		Operation:  OperationBlock,
		Status:     StatusBlocked,
	})
}

func NewInventoryStockReleased(data StockData) Event {
	return newEvent(EventTypeInventoryStockReleased, InventoryStockPayload{
		HotelID:    data.HotelID,
		RoomTypeID: data.RoomTypeID,
		Quantity:   data.Quantity,
		Operation:  OperationRelease,
		Status:     StatusReleased,
	})
}

func NewInventoryStockFailed(data StockData, reason string) Event {
	return newEvent(EventTypeInventoryStockFailed, InventoryStockFailedPayload{
		HotelID:    data.HotelID,
		RoomTypeID: data.RoomTypeID,
		Quantity:   data.Quantity,
		Operation:  data.Operation,
		Status:     StatusFailed,
		Reason:     reason,
	})
}

func newEvent(eventType string, payload interface{}) Event {
	return Event{
		EventID:       uuid.NewString(),
		EventType:     eventType,
		Version:       1,
		SourceService: SourceInventoryService,
		OccurredAt:    time.Now().UTC().Format(time.RFC3339),
		Payload:       payload,
	}
}
