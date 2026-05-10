CREATE TABLE IF NOT EXISTS room_types (
    id UUID PRIMARY KEY,
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    nombre VARCHAR(255) NOT NULL,
    precio_noche DECIMAL(10, 2) NOT NULL,
    capacidad INT NOT NULL,
    stock_total INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_room_types_hotel_id ON room_types(hotel_id);
CREATE INDEX IF NOT EXISTS idx_room_types_capacidad ON room_types(capacidad);
CREATE INDEX IF NOT EXISTS idx_room_types_precio ON room_types(precio_noche);
