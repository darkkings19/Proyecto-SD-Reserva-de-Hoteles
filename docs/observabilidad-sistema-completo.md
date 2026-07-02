# Observabilidad del sistema completo

Esta rama extiende la observabilidad desde el modulo de usuarios hacia los cuatro modulos principales:

- Usuarios: `user-service`
- Reservas: `reservation_service`
- Inventario / Hoteles: `inventario_service`
- Notificaciones: `notification_service`

Usuarios expone metricas con Spring Boot Actuator en `/actuator/prometheus`. Reservas, inventario y notificaciones exponen `/metrics`.

## Endpoints de metricas

| Servicio | Job Prometheus | Endpoint interno | Endpoint host |
| --- | --- | --- | --- |
| Usuarios | `user-service` | `user-service:8080/actuator/prometheus` | `http://localhost:8081/actuator/prometheus` |
| Reservas | `reservation-service` | `reservation_service:9102/metrics` | `http://localhost:9102/metrics` |
| Inventario / Hoteles | `inventory-service` | `inventario_service:9103/metrics` | `http://localhost:9103/metrics` |
| Notificaciones | `notification-service` | `notification_service:9104/metrics` | `http://localhost:9104/metrics` |

## Dashboard Grafana

```text
http://localhost:3001
usuario: admin
password: admin
```

Dashboard:

```text
Origen X / Observabilidad Sistema Completo
```

Se eligio un solo dashboard unificado porque permite mostrar el estado distribuido completo en una sola vista. Esta separado por filas: estado general, usuarios, reservas, inventario/hoteles y notificaciones.

## Metricas por modulo

Usuarios:

- `users_created_total`
- `users_login_success_total`
- `users_login_failed_total`
- `users_active_sessions`
- `users_logout_success_total`
- `users_token_validated_total`
- `users_domain_errors_total`

Reservas:

- `reservations_created_total`
- `reservations_list_total`
- `reservations_get_total`
- `reservations_failures_total{stage="..."}`
- `reservations_notifications_total{status="..."}`
- `reservations_create_duration_seconds`

Inventario / Hoteles:

- `inventory_search_total`
- `inventory_search_empty_total`
- `inventory_stock_updates_total{action="bloquear|liberar"}`
- `inventory_errors_total{operation="...", error_type="..."}`
- `inventory_available_rooms`

Notificaciones:

- `notifications_saved_total{tipo="..."}`
- `notifications_external_total{status="success|failed"}`
- `notifications_errors_total{operation="...", error_type="..."}`
- `notification_send_duration_seconds`

## Como generar datos para la demo

```powershell
docker compose up -d --build
```

Crear usuario, login, buscar habitaciones y crear reserva:

```powershell
$email = "demo$(Get-Random)@test.com"
$user = Invoke-RestMethod -Uri "http://localhost:8080/users" -Method Post -ContentType "application/json" -Body (@{
  nombre = "Usuario Demo"
  email = $email
  password = "pass123"
  telefono = "123456"
} | ConvertTo-Json)

$login = Invoke-RestMethod -Uri "http://localhost:8080/login" -Method Post -ContentType "application/json" -Body (@{
  email = $email
  password = "pass123"
} | ConvertTo-Json)

$inventory = Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/search" -Method Post -ContentType "application/json" -Body (@{
  fecha_inicio = "2026-07-10"
  fecha_fin = "2026-07-12"
  ubicacion = ""
  precio_max = 0
  capacidad = 0
} | ConvertTo-Json)

$room = $inventory.rooms[0]
Invoke-RestMethod -Uri "http://localhost:8080/reservations" -Method Post -ContentType "application/json" -Body (@{
  user_id = $user.id
  hotel_id = $room.hotel_id
  room_type_id = $room.room_type_id
  fecha_inicio = "2026-07-10"
  fecha_fin = "2026-07-12"
} | ConvertTo-Json)
```

## Consultas Prometheus utiles

Estado de los cuatro servicios:

```text
up{job=~"user-service|reservation-service|inventory-service|notification-service"}
```

Eventos por periodo:

```text
round(increase(users_created_total[15m]))
round(increase(reservations_created_total[15m]))
round(increase(inventory_search_total[15m]))
round(sum(increase(notifications_saved_total[15m])))
```

Estado actual:

```text
users_active_sessions
inventory_available_rooms
```

Errores:

```text
sum by (stage) (increase(reservations_failures_total[15m]))
sum by (operation, error_type) (increase(inventory_errors_total[15m]))
sum by (operation, error_type) (increase(notifications_errors_total[15m]))
```

## Trade-offs

- No se agrego JWT obligatorio entre microservicios; el foco de esta rama es observabilidad.
- No se agrego tracing distribuido con OpenTelemetry.
- En notificaciones, el envio externo depende de `RESEND_API_KEY`; sin esa clave se observa el guardado y el flujo interno, no un correo real.
