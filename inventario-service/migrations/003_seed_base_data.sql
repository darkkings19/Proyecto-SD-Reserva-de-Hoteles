-- Seed: Minimal data for automatic initialization via docker-entrypoint-initdb.d
INSERT INTO hotels (id, nombre, ubicacion, caracteristicas) VALUES 
('11111111-1111-1111-1111-111111111111', 'Hotel Origen Centro', 'Santiago', 'Wifi, Piscina, Desayuno incluido, Gimnasio'),
('22222222-2222-2222-2222-222222222222', 'Hotel Montaña Andina', 'Cordillera', 'Wifi, SPA, Naturaleza, Trekking guiado')
ON CONFLICT (id) DO NOTHING;

INSERT INTO room_types (id, hotel_id, nombre, precio_noche, capacidad, stock_total) VALUES 
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'Habitación Simple', 45000.00, 1, 10),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'Habitación Doble', 72000.00, 2, 5),
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222222', 'Cabaña Familiar', 120000.00, 4, 3)
ON CONFLICT (id) DO NOTHING;
