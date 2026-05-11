package service

import (
	"context"

	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/domain"
	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/repository"
	"github.com/google/uuid"
)

type InventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) SearchAvailableRooms(ctx context.Context, location string, maxPrice float64, minCapacity int) ([]domain.RoomAvailability, error) {
	// Se puede agregar lógica de validación adicional si es necesaria.
	return s.repo.SearchAvailableRooms(ctx, location, maxPrice, minCapacity)
}

func (s *InventoryService) UpdateStock(ctx context.Context, roomTypeID string, quantity int, action string) (bool, error) {
	return s.repo.UpdateStock(ctx, roomTypeID, quantity, action)
}

func (s *InventoryService) CreateHotel(ctx context.Context, nombre, ubicacion, caracteristicas string) (*domain.Hotel, error) {
	hotel := &domain.Hotel{
		ID:              uuid.New().String(),
		Nombre:          nombre,
		Ubicacion:       ubicacion,
		Caracteristicas: caracteristicas,
	}

	err := s.repo.CreateHotel(ctx, hotel)
	if err != nil {
		return nil, err
	}

	return hotel, nil
}
