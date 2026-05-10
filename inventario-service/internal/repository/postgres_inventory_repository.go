package repository

import (
	"context"
	"fmt"

	"github.com/darkkings19/Proyecto-SD-Reserva-de-Hoteles/inventario-service/internal/domain"
	"gorm.io/gorm"
)

type postgresInventoryRepository struct {
	db *gorm.DB
}

func NewPostgresInventoryRepository(db *gorm.DB) InventoryRepository {
	return &postgresInventoryRepository{db: db}
}

func (r *postgresInventoryRepository) SearchAvailableRooms(ctx context.Context, location string, maxPrice float64, minCapacity int) ([]domain.RoomAvailability, error) {
	var results []domain.RoomAvailability

	query := r.db.WithContext(ctx).Table("room_types rt").
		Select(`
			h.id as hotel_id, 
			h.nombre as nombre_hotel, 
			rt.id as room_type_id, 
			rt.precio_noche as precio_noche, 
			rt.capacidad as capacidad, 
			rt.stock_total as stock_disponible
		`).
		Joins("JOIN hotels h ON h.id = rt.hotel_id").
		Where("rt.stock_total > 0")

	if location != "" {
		query = query.Where("h.ubicacion ILIKE ?", "%"+location+"%")
	}
	if maxPrice > 0 {
		query = query.Where("rt.precio_noche <= ?", maxPrice)
	}
	if minCapacity > 0 {
		query = query.Where("rt.capacidad >= ?", minCapacity)
	}

	err := query.Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("error querying available rooms: %w", err)
	}

	return results, nil
}

func (r *postgresInventoryRepository) UpdateStock(ctx context.Context, roomTypeID string, quantity int, action string) (bool, error) {
	if quantity <= 0 {
		return false, domain.ErrInvalidQuantity
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		switch action {
		case "BLOQUEAR":
			// Atomic conditional update: UPDATE room_types SET stock_total = stock_total - ? WHERE id = ? AND stock_total >= ?
			// This prevents race conditions at the DB level (§4.3 of Documento Técnico).
			res := tx.Model(&domain.RoomType{}).
				Where("id = ? AND stock_total >= ?", roomTypeID, quantity).
				UpdateColumn("stock_total", gorm.Expr("stock_total - ?", quantity))

			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				// Differentiate between "not found" and "insufficient stock"
				var count int64
				tx.Model(&domain.RoomType{}).Where("id = ?", roomTypeID).Count(&count)
				if count == 0 {
					return domain.ErrRoomTypeNotFound
				}
				return domain.ErrInsufficientStock
			}

		case "LIBERAR":
			res := tx.Model(&domain.RoomType{}).
				Where("id = ?", roomTypeID).
				UpdateColumn("stock_total", gorm.Expr("stock_total + ?", quantity))

			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return domain.ErrRoomTypeNotFound
			}

		default:
			return domain.ErrInvalidAction
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
