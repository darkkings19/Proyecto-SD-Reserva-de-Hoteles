package grpc_test

import (
	"context"
	"testing"

	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/domain"
	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/service"
	handler "github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/transport/grpc"
	pb "github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- Mock Repository ---

type mockRepo struct {
	searchResult []domain.RoomAvailability
	searchErr    error
	updateOk     bool
	updateErr    error
}

func (m *mockRepo) SearchAvailableRooms(_ context.Context, _ string, _ float64, _ int) ([]domain.RoomAvailability, error) {
	return m.searchResult, m.searchErr
}

func (m *mockRepo) UpdateStock(_ context.Context, _ string, _ int, _ string) (bool, error) {
	return m.updateOk, m.updateErr
}

func (m *mockRepo) CreateHotel(_ context.Context, _ *domain.Hotel) error {
	return m.updateErr
}

// --- Handler Tests ---

func TestHandler_SearchAvailableRooms_OK(t *testing.T) {
	repo := &mockRepo{
		searchResult: []domain.RoomAvailability{
			{HotelID: "h1", NombreHotel: "Test", RoomTypeID: "rt1", PrecioNoche: 50000, Capacidad: 2, StockDisponible: 5},
		},
	}
	svc := service.NewInventoryService(repo)
	h := handler.NewInventoryHandler(svc)

	resp, err := h.SearchAvailableRooms(context.Background(), &pb.SearchRequest{
		Ubicacion: "Santiago",
		Capacidad: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(resp.Rooms))
	}
	if resp.Rooms[0].StockDisponible != 5 {
		t.Errorf("expected stock_disponible=5, got %d", resp.Rooms[0].StockDisponible)
	}
}

func TestHandler_UpdateStock_ResourceExhausted(t *testing.T) {
	repo := &mockRepo{updateOk: false, updateErr: domain.ErrInsufficientStock}
	svc := service.NewInventoryService(repo)
	h := handler.NewInventoryHandler(svc)

	_, err := h.UpdateStock(context.Background(), &pb.UpdateStockRequest{
		RoomTypeId: "rt1",
		Cantidad:   1,
		Accion:     pb.Action_BLOQUEAR,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T", err)
	}
	if st.Code() != codes.ResourceExhausted {
		t.Errorf("expected ResourceExhausted, got %v", st.Code())
	}
}

func TestHandler_UpdateStock_NotFound(t *testing.T) {
	repo := &mockRepo{updateOk: false, updateErr: domain.ErrRoomTypeNotFound}
	svc := service.NewInventoryService(repo)
	h := handler.NewInventoryHandler(svc)

	_, err := h.UpdateStock(context.Background(), &pb.UpdateStockRequest{
		RoomTypeId: "nonexistent",
		Cantidad:   1,
		Accion:     pb.Action_LIBERAR,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %T", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %v", st.Code())
	}
}

func TestHandler_UpdateStock_Success(t *testing.T) {
	repo := &mockRepo{updateOk: true}
	svc := service.NewInventoryService(repo)
	h := handler.NewInventoryHandler(svc)

	resp, err := h.UpdateStock(context.Background(), &pb.UpdateStockRequest{
		RoomTypeId: "rt1",
		Cantidad:   1,
		Accion:     pb.Action_BLOQUEAR,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
}

func TestHandler_CreateHotel_Success(t *testing.T) {
	repo := &mockRepo{updateOk: true}
	svc := service.NewInventoryService(repo)
	h := handler.NewInventoryHandler(svc)

	resp, err := h.CreateHotel(context.Background(), &pb.CreateHotelRequest{
		Nombre: "Test", Ubicacion: "Lugar", Caracteristicas: "X",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Id == "" {
		t.Fatal("expected id")
	}
	if resp.Nombre != "Test" {
		t.Errorf("expected Test, got %s", resp.Nombre)
	}
}
