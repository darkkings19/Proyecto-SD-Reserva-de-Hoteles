# Kafka Demo Guide - Origen X

## 1. Levantar stack

```bash
docker compose up -d
docker compose ps
```

Frontend:

```text
http://localhost:3000
```

Kafka UI:

```text
http://localhost:8082
```

## 2. Abrir consumers

Usar una terminal por topico en Git Bash.

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.gateway.events --from-beginning'
```

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.users.events --from-beginning'
```

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.reservations.events --from-beginning'
```

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.inventory.events --from-beginning'
```

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.notifications.events --from-beginning'
```

## 3. Prueba end-to-end esperada

| Accion | Eventos esperados |
|---|---|
| Crear usuario | `GatewayUserRegistered`, `UserCreated` |
| Iniciar sesion | `GatewayUserLoggedIn`, `UserLoggedIn` |
| Buscar habitaciones | `GatewayInventorySearchRequested` |
| Crear reserva | `GatewayReservationRequested`, `ReservationCreated`, `InventoryStockBlocked`, `ReservationConfirmed`, `NotificationSent` |
| Editar perfil | `UserUpdated` |

## 4. Guion de demo

1. Mostrar `docker compose ps` con servicios arriba.
2. Abrir Kafka UI y mostrar topicos `origenx.*`.
3. Abrir el frontend.
4. Crear un usuario y mostrar eventos en `gateway` y `users`.
5. Iniciar sesion y mostrar eventos de login.
6. Buscar habitaciones y mostrar `GatewayInventorySearchRequested`.
7. Crear reserva y mostrar la cadena de eventos de gateway, reservas, inventario y notificaciones.
8. Editar perfil y mostrar `UserUpdated`.

## 5. Troubleshooting

| Problema | Solucion |
|---|---|
| Puerto `8080` ocupado | Cerrar el proceso que lo usa o cambiar el puerto publicado de `api_gateway`. |
| Git Bash convierte `/opt/kafka` | Usar `docker compose exec kafka sh -lc '...'`. |
| Kafka no esta healthy | Esperar y revisar `docker compose ps kafka`. |
| `UNKNOWN_TOPIC_OR_PARTITION` | Ejecutar primero una accion que publique en ese topico o revisar Kafka UI. |
| Warnings `RESEND_API_KEY` y `USER_SERVICE_HOST` | Son warnings de entorno; no impiden la demo Kafka. |
| Consumer no muestra eventos antiguos | Confirmar que el topico tenga mensajes; usar `--from-beginning` y revisar que se esta mirando el topico correcto. |
