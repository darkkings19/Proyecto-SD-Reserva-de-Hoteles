package repository

import (
	"context"

	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/domain"
)

type InventoryRepository interface {
	SearchAvailableRooms(ctx context.Context, location string, maxPrice float64, minCapacity int) ([]domain.RoomAvailability, error)
	UpdateStock(ctx context.Context, roomTypeID string, quantity int, action string) (bool, error)
	CreateHotel(ctx context.Context, hotel *domain.Hotel) error
}
