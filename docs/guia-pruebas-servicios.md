# Guia de pruebas del sistema

Esta guia sirve para probar el proyecto completo desde consola, desde el frontend y desde la visualizacion de observabilidad.

## 1. Levantar el sistema

Desde la raiz del proyecto:

```powershell
docker compose up -d --build
```

Verificar contenedores:

```powershell
docker compose ps
```

Servicios principales:

| Servicio | URL o puerto | Uso |
| --- | --- | --- |
| Frontend | `http://localhost:3000` | Interfaz visual de usuarios, inventario y reservas |
| API Gateway | `http://localhost:8080` | Entrada HTTP para probar el sistema |
| User Service gRPC | `localhost:9090` | Servicio interno de usuarios |
| User Service Actuator | `http://localhost:8081` | Health y metricas del modulo usuarios |
| Prometheus | `http://localhost:9091` | Consultas PromQL |
| Grafana | `http://localhost:3001` | Dashboard visual |
| Reservation Service gRPC | `localhost:50052` | Servicio interno de reservas |
| Inventory Service gRPC | `localhost:50053` | Servicio interno de inventario |

## 2. Health checks basicos

API Gateway:

```powershell
Invoke-RestMethod http://localhost:8080/health
```

User Service:

```powershell
Invoke-RestMethod http://localhost:8081/actuator/health | ConvertTo-Json -Depth 8
```

Prometheus recolectando `user-service`:

```powershell
Invoke-RestMethod "http://localhost:9091/api/v1/query?query=up%7Bjob%3D%22user-service%22%7D" | ConvertTo-Json -Depth 8
```

Grafana:

```powershell
$basic = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("admin:admin"))
Invoke-RestMethod -Headers @{Authorization="Basic $basic"} http://localhost:3001/api/health
```

## 3. Probar usuarios por consola

Crear usuario:

```powershell
$email = "demo$(Get-Random)@test.com"
$user = Invoke-RestMethod -Uri "http://localhost:8080/users" -Method Post -ContentType "application/json" -Body (@{
  nombre = "Usuario Demo"
  email = $email
  password = "pass123"
  telefono = "123456"
} | ConvertTo-Json)
$user
```

Login exitoso:

```powershell
$login = Invoke-RestMethod -Uri "http://localhost:8080/login" -Method Post -ContentType "application/json" -Body (@{
  email = $email
  password = "pass123"
} | ConvertTo-Json)
$login.access_token
```

Validar token:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/validate-token" -Method Post -ContentType "application/json" -Body (@{
  access_token = $login.access_token
} | ConvertTo-Json)
```

Login fallido:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/login" -Method Post -ContentType "application/json" -Body (@{
  email = $email
  password = "incorrecta"
} | ConvertTo-Json)
```

Ese comando debe responder `401`. Sirve para mostrar `users_login_failed_total`.

Logout:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/logout" -Method Post -ContentType "application/json" -Body (@{
  access_token = $login.access_token
} | ConvertTo-Json)
```

Despues del logout, validar el mismo token debe responder `401`.

## 4. Probar inventario por consola

Buscar habitaciones disponibles:

```powershell
$inventory = Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/search" -Method Post -ContentType "application/json" -Body (@{
  fecha_inicio = "2026-07-10"
  fecha_fin = "2026-07-12"
  ubicacion = ""
  precio_max = 0
  capacidad = 0
} | ConvertTo-Json)
$inventory.rooms
```

Datos base esperados:

```text
hotel_id: h1
room_type_id: rt1
nombre_hotel: Hotel Continental
room_type_name: Suite Ejecutiva
```

## 5. Probar reservas por consola

Crear reserva usando el usuario creado y la primera habitacion disponible:

```powershell
$room = $inventory.rooms[0]
$reservation = Invoke-RestMethod -Uri "http://localhost:8080/reservations" -Method Post -ContentType "application/json" -Body (@{
  user_id = $user.id
  hotel_id = $room.hotel_id
  room_type_id = $room.room_type_id
  fecha_inicio = "2026-07-10"
  fecha_fin = "2026-07-12"
} | ConvertTo-Json)
$reservation
```

Listar reservas:

```powershell
Invoke-RestMethod http://localhost:8080/reservations
```

Este flujo prueba varios servicios juntos:

- API Gateway recibe HTTP.
- Reservation Service crea la reserva.
- Reservation Service consulta User Service para validar usuario.
- Reservation Service llama Inventory Service para bloquear stock.
- Reservation Service llama Notification Service para registrar/enviar confirmacion.

## 6. Probar notificaciones

El servicio de notificaciones no tiene endpoint HTTP directo en el gateway. Se prueba indirectamente creando una reserva.

Ver logs:

```powershell
docker logs notification_service --tail 80
```

Ver base de datos de notificaciones:

```powershell
docker exec -it notificaciones_db psql -U postgres -d notificaciones_db -c "SELECT user_id, reservation_id, tipo, email, created_at FROM notifications ORDER BY created_at DESC LIMIT 5;"
```

Si `RESEND_API_KEY` esta vacio, puede no enviarse correo real, pero el flujo interno y el registro en base de datos sirven para demostrar integracion.

## 7. Probar desde el frontend

Abrir:

```text
http://localhost:3000
```

Pasos:

1. Entrar a la pestana `Registrarse`.
2. Crear un usuario.
3. Copiar o revisar que el ID se ponga automaticamente en `Usuario ID`.
4. Entrar con email y password.
5. Seleccionar hotel y tipo de habitacion.
6. Crear reserva.
7. Ver la reserva en `Mis Reservas`.
8. Presionar `Salir` para ejecutar logout real contra el backend.

Importante: el logout visual ahora llama al endpoint `/logout`, por eso la metrica `users_active_sessions` baja despues de cerrar sesion.

## 8. Probar observabilidad

Prometheus:

```text
http://localhost:9091
```

Consultas utiles:

```text
up{job="user-service"}
users_created_total
users_login_success_total
users_login_failed_total
users_active_sessions
users_logout_success_total
users_token_validated_total
sum(users_domain_errors_total)
```

Metricas por periodo:

```text
round(increase(users_created_total[15m]))
round(increase(users_login_success_total[15m]))
round(increase(users_login_failed_total[15m]))
```

Grafana:

```text
http://localhost:3001
usuario: admin
password: admin
```

Abrir:

```text
Origen X / Observabilidad User Service
```

Dashboard completo:

```text
Origen X / Observabilidad Sistema Completo
```

Que mostrar:

- Estado del `user-service`.
- Usuarios creados total.
- Logins exitosos total.
- Logins fallidos total.
- Usuarios creados en el periodo seleccionado.
- Logins exitosos en el periodo seleccionado.
- Logins fallidos en el periodo seleccionado.
- Usuarios activos ahora.
- Reservas creadas, fallos por etapa, busquedas de inventario, stock disponible y notificaciones guardadas en el dashboard completo.

El selector de tiempo de Grafana esta arriba a la derecha. Ahi se puede elegir `Last 5 minutes`, `Today` o un rango personalizado con dia y hora.

## 9. Logs utiles

```powershell
docker logs user-service --tail 80
docker logs api_gateway --tail 80
docker logs reservation_service --tail 80
docker logs inventario_service --tail 80
docker logs notification_service --tail 80
docker logs prometheus --tail 80
docker logs grafana --tail 80
```

## 10. Pruebas automaticas

User Service:

```powershell
cd user-service
mvn test
```

Inventory Service:

```powershell
cd inventario-service
go test ./...
```

Notification Service:

```powershell
cd notification_service
python -m pytest
```

Reservation Service:

```powershell
cd mi-servicio
go test ./...
```
