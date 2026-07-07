package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeReservationCreated   = "ReservationCreated"
	EventTypeReservationConfirmed = "ReservationConfirmed"
	EventTypeReservationFailed    = "ReservationFailed"

	SourceReservationService = "reservation-service"
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

type ReservationCreatedPayload struct {
	UserID      string `json:"user_id"`
	HotelID     string `json:"hotel_id"`
	RoomTypeID  string `json:"room_type_id"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
	Status      string `json:"status"`
}

type ReservationConfirmedPayload struct {
	ReservationID string  `json:"reservation_id"`
	UserID        string  `json:"user_id"`
	HotelID       string  `json:"hotel_id"`
	RoomTypeID    string  `json:"room_type_id"`
	FechaInicio   string  `json:"fecha_inicio"`
	FechaFin      string  `json:"fecha_fin"`
	Status        string  `json:"status"`
	MontoTotal    float64 `json:"monto_total"`
}

type ReservationFailedPayload struct {
	UserID      string `json:"user_id"`
	HotelID     string `json:"hotel_id"`
	RoomTypeID  string `json:"room_type_id"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

type ReservationData struct {
	ReservationID string
	UserID        string
	HotelID       string
	RoomTypeID    string
	FechaInicio   string
	FechaFin      string
	MontoTotal    float64
}

func NewReservationCreated(data ReservationData) Event {
	return newEvent(EventTypeReservationCreated, ReservationCreatedPayload{
		UserID:      data.UserID,
		HotelID:     data.HotelID,
		RoomTypeID:  data.RoomTypeID,
		FechaInicio: data.FechaInicio,
		FechaFin:    data.FechaFin,
		Status:      "CREATING",
	})
}

func NewReservationConfirmed(data ReservationData) Event {
	return newEvent(EventTypeReservationConfirmed, ReservationConfirmedPayload{
		ReservationID: data.ReservationID,
		UserID:        data.UserID,
		HotelID:       data.HotelID,
		RoomTypeID:    data.RoomTypeID,
		FechaInicio:   data.FechaInicio,
		FechaFin:      data.FechaFin,
		Status:        "CONFIRMED",
		MontoTotal:    data.MontoTotal,
	})
}

func NewReservationFailed(data ReservationData, reason string) Event {
	return newEvent(EventTypeReservationFailed, ReservationFailedPayload{
		UserID:      data.UserID,
		HotelID:     data.HotelID,
		RoomTypeID:  data.RoomTypeID,
		FechaInicio: data.FechaInicio,
		FechaFin:    data.FechaFin,
		Status:      "FAILED",
		Reason:      reason,
	})
}

func newEvent(eventType string, payload interface{}) Event {
	return Event{
		EventID:       uuid.NewString(),
		EventType:     eventType,
		Version:       1,
		SourceService: SourceReservationService,
		OccurredAt:    time.Now().UTC().Format(time.RFC3339),
		Payload:       payload,
	}
}
