package service_test

import (
	"context"
	"testing"

	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/domain"
	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/service"
)

// --- Mock Repository ---

type mockInventoryRepository struct {
	searchResult []domain.RoomAvailability
	searchErr    error
	updateResult bool
	updateErr    error
	lastAction   string
	lastRoomID   string
	lastQty      int
}

func (m *mockInventoryRepository) SearchAvailableRooms(ctx context.Context, location string, maxPrice float64, minCapacity int) ([]domain.RoomAvailability, error) {
	return m.searchResult, m.searchErr
}

func (m *mockInventoryRepository) UpdateStock(ctx context.Context, roomTypeID string, quantity int, action string) (bool, error) {
	m.lastRoomID = roomTypeID
	m.lastQty = quantity
	m.lastAction = action
	return m.updateResult, m.updateErr
}

// --- Tests ---

func TestSearchAvailableRooms_ReturnsResults(t *testing.T) {
	mock := &mockInventoryRepository{
		searchResult: []domain.RoomAvailability{
			{HotelID: "h1", NombreHotel: "Hotel Test", RoomTypeID: "rt1", PrecioNoche: 50000, Capacidad: 2, StockDisponible: 5},
		},
	}
	svc := service.NewInventoryService(mock)

	rooms, err := svc.SearchAvailableRooms(context.Background(), "Santiago", 100000, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(rooms))
	}
	if rooms[0].HotelID != "h1" {
		t.Errorf("expected hotel_id h1, got %s", rooms[0].HotelID)
	}
}

func TestSearchAvailableRooms_ReturnsEmpty(t *testing.T) {
	mock := &mockInventoryRepository{
		searchResult: []domain.RoomAvailability{},
	}
	svc := service.NewInventoryService(mock)

	rooms, err := svc.SearchAvailableRooms(context.Background(), "NoExiste", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 0 {
		t.Fatalf("expected 0 rooms, got %d", len(rooms))
	}
}

func TestUpdateStock_BloquearSuccess(t *testing.T) {
	mock := &mockInventoryRepository{updateResult: true}
	svc := service.NewInventoryService(mock)

	ok, err := svc.UpdateStock(context.Background(), "rt1", 1, "BLOQUEAR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected true")
	}
	if mock.lastAction != "BLOQUEAR" {
		t.Errorf("expected action BLOQUEAR, got %s", mock.lastAction)
	}
}

func TestUpdateStock_InsufficientStock(t *testing.T) {
	mock := &mockInventoryRepository{
		updateResult: false,
		updateErr:    domain.ErrInsufficientStock,
	}
	svc := service.NewInventoryService(mock)

	ok, err := svc.UpdateStock(context.Background(), "rt1", 1, "BLOQUEAR")
	if ok {
		t.Fatal("expected false when stock is insufficient")
	}
	if err != domain.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestUpdateStock_RoomTypeNotFound(t *testing.T) {
	mock := &mockInventoryRepository{
		updateResult: false,
		updateErr:    domain.ErrRoomTypeNotFound,
	}
	svc := service.NewInventoryService(mock)

	ok, err := svc.UpdateStock(context.Background(), "nonexistent", 1, "LIBERAR")
	if ok {
		t.Fatal("expected false when room type not found")
	}
	if err != domain.ErrRoomTypeNotFound {
		t.Errorf("expected ErrRoomTypeNotFound, got %v", err)
	}
}

func TestUpdateStock_LiberarSuccess(t *testing.T) {
	mock := &mockInventoryRepository{updateResult: true}
	svc := service.NewInventoryService(mock)

	ok, err := svc.UpdateStock(context.Background(), "rt1", 1, "LIBERAR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected true")
	}
	if mock.lastAction != "LIBERAR" {
		t.Errorf("expected action LIBERAR, got %s", mock.lastAction)
	}
}
