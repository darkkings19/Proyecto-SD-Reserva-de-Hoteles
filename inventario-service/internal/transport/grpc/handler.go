package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/darkkings19/inventario-service/internal/domain"
	"github.com/darkkings19/inventario-service/internal/service"
	pb "github.com/darkkings19/inventario-service/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultTimeout = 5 * time.Second

// InventoryHandler implements the gRPC InventoryServiceServer interface.
type InventoryHandler struct {
	pb.UnimplementedInventoryServiceServer
	svc *service.InventoryService
}

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) SearchAvailableRooms(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	// §4.4 Resiliencia mediante Timeouts Contextuales
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	rooms, err := h.svc.SearchAvailableRooms(ctx, req.Ubicacion, req.PrecioMax, int(req.Capacidad))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search rooms: %v", err)
	}

	var pbRooms []*pb.RoomTypeAvailability
	for _, r := range rooms {
		pbRooms = append(pbRooms, &pb.RoomTypeAvailability{
			HotelId:         r.HotelID,
			NombreHotel:     r.NombreHotel,
			RoomTypeId:      r.RoomTypeID,
			PrecioNoche:     r.PrecioNoche,
			Capacidad:       int32(r.Capacidad),
			StockDisponible: int32(r.StockDisponible),
		})
	}

	return &pb.SearchResponse{Rooms: pbRooms}, nil
}

func (h *InventoryHandler) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	// §4.4 Resiliencia mediante Timeouts Contextuales
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	actionStr := "BLOQUEAR"
	if req.Accion == pb.Action_LIBERAR {
		actionStr = "LIBERAR"
	}

	success, err := h.svc.UpdateStock(ctx, req.RoomTypeId, int(req.Cantidad), actionStr)
	if err != nil {
		// Map domain sentinel errors to proper gRPC status codes
		if errors.Is(err, domain.ErrInsufficientStock) {
			return nil, status.Errorf(codes.ResourceExhausted, "insufficient stock for room_type_id %s", req.RoomTypeId)
		}
		if errors.Is(err, domain.ErrRoomTypeNotFound) {
			return nil, status.Errorf(codes.NotFound, "room_type_id %s not found", req.RoomTypeId)
		}
		if errors.Is(err, domain.ErrInvalidQuantity) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid quantity: %d", req.Cantidad)
		}
		return nil, status.Errorf(codes.Internal, "failed to update stock: %v", err)
	}

	return &pb.UpdateStockResponse{Status: success}, nil
}
