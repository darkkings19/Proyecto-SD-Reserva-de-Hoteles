-- Seed: Demo data for development and testing
-- This file is NOT auto-loaded by docker-entrypoint-initdb.d.
-- Use it manually: psql -h localhost -U postgres -d inventario_db -f seed/demo_data.sql

-- Hotels
INSERT INTO hotels (id, nombre, ubicacion, caracteristicas) VALUES
('11111111-1111-1111-1111-111111111111', 'Hotel Origen Centro', 'Santiago', 'Wifi, Piscina, Desayuno incluido, Gimnasio'),
('22222222-2222-2222-2222-222222222222', 'Hotel Montaña Andina', 'Cordillera', 'Wifi, SPA, Naturaleza, Trekking guiado'),
('33333333-3333-3333-3333-333333333333', 'Hotel Playa Sur', 'Valparaíso', 'Wifi, Vista al mar, Bar, Restaurante')
ON CONFLICT (id) DO NOTHING;

-- Room Types
INSERT INTO room_types (id, hotel_id, nombre, precio_noche, capacidad, stock_total) VALUES
-- Hotel Origen Centro
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'Habitación Simple', 45000.00, 1, 15),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'Habitación Doble',  72000.00, 2, 10),
('cccccccc-cccc-cccc-cccc-cccccccccccc', '11111111-1111-1111-1111-111111111111', 'Suite Premium',     150000.00, 2, 3),
-- Hotel Montaña Andina
('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222222', 'Cabaña Familiar',   120000.00, 4, 5),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '22222222-2222-2222-2222-222222222222', 'Habitación Doble',  65000.00, 2, 8),
-- Hotel Playa Sur
('ffffffff-ffff-ffff-ffff-ffffffffffff', '33333333-3333-3333-3333-333333333333', 'Habitación Vista Mar', 95000.00, 2, 6),
('00000000-0000-0000-0000-000000000001', '33333333-3333-3333-3333-333333333333', 'Habitación Simple',    40000.00, 1, 12)
ON CONFLICT (id) DO NOTHING;
