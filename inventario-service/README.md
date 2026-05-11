# Inventario Service — Origen X

Microservicio de **Inventario/Hoteles** del sistema Origen X. Gestiona hoteles, tipos de habitaciones, precios y disponibilidad de stock. Comunicación exclusivamente por **gRPC** dentro de la red interna `origen-net`.

---

## Arquitectura

| Aspecto | Detalle |
|---|---|
| **Lenguaje** | Go 1.22 |
| **Base de Datos** | PostgreSQL 15 (`inventario_db`) — independiente por servicio |
| **Comunicación** | gRPC + Protobuf (sin REST) |
| **ORM** | GORM con driver pgx |
| **Red** | Solo accesible dentro de `origen-net` — no expone puertos al host |

### Estructura del Proyecto

```
inventario-service/
├── cmd/server/main.go              # Entrypoint del servidor gRPC
├── internal/
│   ├── config/config.go            # Configuración centralizada desde env vars
│   ├── db/db.go                    # Conexión a PostgreSQL (GORM)
│   ├── domain/
│   │   ├── hotel.go                # Entidad Hotel
│   │   ├── room_type.go            # Entidad RoomType + DTO RoomAvailability
│   │   └── errors.go               # Errores centinela del dominio
│   ├── repository/
│   │   ├── repository.go           # Interfaz InventoryRepository
│   │   └── postgres_inventory_repository.go  # Implementación PostgreSQL
│   ├── service/
│   │   ├── inventory_service.go    # Lógica de negocio
│   │   └── inventory_service_test.go
│   └── transport/grpc/
│       ├── handler.go              # Handler gRPC (mapea proto ↔ dominio)
│       └── handler_test.go
├── proto/
│   ├── inventory.proto             # Contrato Protobuf
│   └── gen/                        # Stubs generados (*.pb.go)
├── migrations/                     # SQL de inicialización (docker-entrypoint-initdb.d)
├── seed/demo_data.sql              # Datos extendidos para desarrollo
├── Dockerfile                      # Multi-stage build (golang → alpine)
├── Makefile
├── .env.example
└── README.md
```

---

## Decisiones Técnicas

### 1. `stock_total` como disponibilidad operativa
El campo `stock_total` representa las habitaciones **actualmente disponibles**. Al ejecutar `BLOQUEAR`, se decrementa atómicamente (`UPDATE ... SET stock_total = stock_total - 1 WHERE stock_total >= 1`). Al ejecutar `LIBERAR`, se incrementa. Esta decisión simplifica el modelo para esta entrega sin necesidad de una tabla separada de reservas en Inventario.

### 2. Errores centinela (sentinel errors)
Los errores de dominio (`ErrInsufficientStock`, `ErrRoomTypeNotFound`, etc.) están en `domain/errors.go` y se comparan con `errors.Is()`, evitando comparación de strings frágil.

### 3. `context.WithTimeout` (§4.4 del Documento Técnico)
Cada handler gRPC aplica un timeout de 5 segundos para prevenir goroutine leaks y cascading failures.

### 4. Naming: `stock_disponible`
La respuesta de `SearchAvailableRooms` devuelve el campo como `stock_disponible` tal como indica el Documento Técnico.

---

## Variables de Entorno

| Variable | Default | Descripción |
|---|---|---|
| `DB_HOST` | `localhost` | Host de PostgreSQL |
| `DB_USER` | `postgres` | Usuario |
| `DB_PASSWORD` | `postgres` | Contraseña |
| `DB_NAME` | `inventario_db` | Nombre de la BD |
| `DB_PORT` | `5432` | Puerto |
| `PORT` | `50051` | Puerto gRPC |

---

## Cómo Ejecutar

### Con Docker (recomendado)

Desde la raíz del proyecto (`Proyecto-SD-Reserva-de-Hoteles/`):

```bash
docker compose up -d --build
```

Esto levanta:
- `inventario-db`: PostgreSQL con migraciones y seed automáticos
- `inventario-service`: Servidor gRPC en puerto 50051 (solo red interna)

### Local (desarrollo)

```bash
cd inventario-service
cp .env.example .env
# Editar .env con datos de tu PostgreSQL local
make build
make run
```

---

## Migraciones

Las migraciones están en `migrations/` y se ejecutan automáticamente al inicializar el contenedor de PostgreSQL (`docker-entrypoint-initdb.d`):

1. `001_create_hotels.sql` — Tabla `hotels` con índice en `ubicacion`
2. `002_create_room_types.sql` — Tabla `room_types` con índices en `hotel_id`, `capacidad`, `precio_noche`
3. `003_seed_base_data.sql` — Datos de prueba (2 hoteles, 3 tipos de habitación)

Para datos extendidos: `seed/demo_data.sql` (ejecutar manualmente).

---

## Generar Protobuf

Si modificas `proto/inventory.proto`:

```bash
make proto
```

O manualmente:
```bash
protoc --go_out=proto/gen --go_opt=paths=source_relative \
    --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
    -I=proto proto/inventory.proto
```

Requisitos: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`.

---

## Tests

```bash
make test
```

Tests incluidos:
- **Service layer** (7 tests): búsqueda con/sin resultados, bloquear OK, bloquear sin stock, liberar OK, room type no encontrado, crear hotel OK
- **Handler layer** (5 tests): mapeo correcto a códigos gRPC (`ResourceExhausted`, `NotFound`, success), crear hotel OK

---

## Probar con grpcurl

Desde un contenedor en la misma red Docker:

```bash
docker run --rm -it --network origen-net fullstorydev/grpcurl \
  -plaintext inventario-service:50051 list
```

### 1. Buscar disponibilidad

```bash
docker run --rm -it --network origen-net fullstorydev/grpcurl \
  -plaintext -d '{"ubicacion": "Santiago", "capacidad": 1}' \
  inventario-service:50051 inventory.InventoryService/SearchAvailableRooms
```

**Respuesta esperada:**
```json
{
  "rooms": [
    {
      "hotelId": "11111111-1111-1111-1111-111111111111",
      "nombreHotel": "Hotel Origen Centro",
      "roomTypeId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "precioNoche": 45000,
      "capacidad": 1,
      "stockDisponible": 10
    },
    {
      "hotelId": "11111111-1111-1111-1111-111111111111",
      "nombreHotel": "Hotel Origen Centro",
      "roomTypeId": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
      "precioNoche": 72000,
      "capacidad": 2,
      "stockDisponible": 5
    }
  ]
}
```

### 2. Bloquear stock (Reserva)

```bash
docker run --rm -it --network origen-net fullstorydev/grpcurl \
  -plaintext -d '{"room_type_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "cantidad": 1, "accion": 0}' \
  inventario-service:50051 inventory.InventoryService/UpdateStock
```

### 3. Liberar stock (Cancelación)

```bash
docker run --rm -it --network origen-net fullstorydev/grpcurl \
  -plaintext -d '{"room_type_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "cantidad": 1, "accion": 1}' \
  inventario-service:50051 inventory.InventoryService/UpdateStock
```

### 4. Bloquear sin stock (error esperado)

```bash
docker run --rm -it --network origen-net fullstorydev/grpcurl \
  -plaintext -d '{"room_type_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "cantidad": 999, "accion": 0}' \
  inventario-service:50051 inventory.InventoryService/UpdateStock
```

**Respuesta esperada:** Error gRPC `RESOURCE_EXHAUSTED`.

### 5. Crear Hotel

```bash
docker run --rm -it --network origen-net fullstorydev/grpcurl \
  -plaintext -d '{"nombre": "Hotel VIP", "ubicacion": "Viña del Mar", "caracteristicas": "Frente al mar"}' \
  inventario-service:50051 inventory.InventoryService/CreateHotel
```

**Respuesta esperada:**
```json
{
  "id": "e44d32bb-92ac-4158-b649-166299d636cd",
  "nombre": "Hotel VIP",
  "ubicacion": "Viña del Mar",
  "caracteristicas": "Frente al mar"
}
```

---

## Contratos gRPC

### SearchAvailableRooms

| Campo | Tipo | Descripción |
|---|---|---|
| `fecha_inicio` | string | Formato YYYY-MM-DD (reservado para integración futura) |
| `fecha_fin` | string | Formato YYYY-MM-DD (reservado para integración futura) |
| `ubicacion` | string | Búsqueda parcial por ciudad (ILIKE) |
| `precio_max` | double | Filtro de precio máximo (0 = sin filtro) |
| `capacidad` | int32 | Capacidad mínima (0 = sin filtro) |

### UpdateStock

| Campo | Tipo | Descripción |
|---|---|---|
| `room_type_id` | string | UUID del tipo de habitación |
| `cantidad` | int32 | Unidades a bloquear/liberar |
| `accion` | enum | `BLOQUEAR` (0) o `LIBERAR` (1) |

### CreateHotel

| Campo | Tipo | Descripción |
|---|---|---|
| `nombre` | string | Nombre del hotel |
| `ubicacion` | string | Ubicación del hotel |
| `caracteristicas` | string | Características del hotel |

### Códigos de Error gRPC

| Código | Situación |
|---|---|
| `RESOURCE_EXHAUSTED` | Stock insuficiente al bloquear |
| `NOT_FOUND` | `room_type_id` inexistente |
| `INVALID_ARGUMENT` | Cantidad inválida (≤ 0) |
| `INTERNAL` | Error de base de datos u otro |