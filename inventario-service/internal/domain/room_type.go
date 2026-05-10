package domain

import (
	"time"
)

// RoomType represents a type of room offered by a hotel.
// stock_total represents the current operational availability: it is decremented on BLOQUEAR
// and incremented on LIBERAR. This means stock_total == available rooms right now.
type RoomType struct {
	ID          string    `gorm:"type:uuid;primary_key"`
	HotelID     string    `gorm:"type:uuid;not null"`
	Nombre      string    `gorm:"type:varchar(255);not null"`
	PrecioNoche float64   `gorm:"type:decimal(10,2);not null"`
	Capacidad   int       `gorm:"type:int;not null"`
	StockTotal  int       `gorm:"type:int;not null;default:0"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName overrides GORM's default table naming to match the migration schema.
func (RoomType) TableName() string {
	return "room_types"
}

// RoomAvailability is a read-only DTO for search results.
// It carries data from the JOIN between Hotels and RoomTypes.
type RoomAvailability struct {
	HotelID          string  `gorm:"column:hotel_id"`
	NombreHotel      string  `gorm:"column:nombre_hotel"`
	RoomTypeID       string  `gorm:"column:room_type_id"`
	PrecioNoche      float64 `gorm:"column:precio_noche"`
	Capacidad        int     `gorm:"column:capacidad"`
	StockDisponible  int     `gorm:"column:stock_disponible"`
}
