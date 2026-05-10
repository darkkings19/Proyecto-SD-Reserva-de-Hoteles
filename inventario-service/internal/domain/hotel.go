package domain

import (
	"time"
)

// Hotel represents a hotel entity in the inventory domain.
type Hotel struct {
	ID              string    `gorm:"type:uuid;primary_key"`
	Nombre          string    `gorm:"type:varchar(255);not null"`
	Ubicacion       string    `gorm:"type:varchar(255);not null"`
	Caracteristicas string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName overrides GORM's default table naming to match the migration schema.
func (Hotel) TableName() string {
	return "hotels"
}
