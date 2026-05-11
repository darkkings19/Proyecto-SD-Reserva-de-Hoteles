package domain

import "errors"

// Sentinel errors for the inventory domain.
// Using typed errors instead of string comparison for reliable error handling across layers.
var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrRoomTypeNotFound  = errors.New("room_type_id not found")
	ErrInvalidAction     = errors.New("invalid action: must be BLOQUEAR or LIBERAR")
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
)
