# Contratos de Eventos Kafka - Origen X

## Proposito

Este directorio define los contratos JSON de eventos Kafka para Origen X.

Kafka funciona como un **bus de eventos transversal** para publicar hechos relevantes del sistema de forma asincronica. Su objetivo es apoyar casos como auditoria, trazabilidad, notificaciones, integraciones futuras, analitica y desacoplamiento progresivo entre servicios.

La comunicacion **gRPC se mantiene como el mecanismo principal para operaciones sincronicas** entre microservicios. Por ejemplo, crear una reserva, validar un usuario o bloquear stock siguen ocurriendo por gRPC en el flujo actual. Kafka no reemplaza ese flujo; solo agrega una via asincronica para eventos.

En esta fase, estos contratos son solo documentacion. No existen productores ni consumidores implementados todavia.

## Sobre comun de evento

Todos los eventos publicados en Kafka deben usar el mismo sobre base:

```json
{
  "event_id": "uuid",
  "event_type": "ReservationConfirmed",
  "version": 1,
  "source_service": "reservation-service",
  "occurred_at": "2026-07-05T16:00:00Z",
  "correlation_id": "uuid-opcional",
  "payload": {}
}
```

Campos obligatorios:

| Campo | Tipo | Descripcion |
|---|---|---|
| `event_id` | string UUID | Identificador unico del evento. Se usa para idempotencia y deduplicacion. |
| `event_type` | string | Nombre logico del evento, por ejemplo `UserCreated`. |
| `version` | number | Version del contrato del evento. |
| `source_service` | string | Servicio que publica el evento. |
| `occurred_at` | string ISO-8601 UTC | Fecha y hora real en que ocurrio el hecho de negocio. |
| `correlation_id` | string UUID o null | Identificador opcional para correlacionar una operacion distribuida. |
| `payload` | object | Datos especificos del evento. |

## Topicos definidos

| Topico | Proposito |
|---|---|
| `origenx.users.events` | Eventos relacionados con usuarios y perfiles. |
| `origenx.reservations.events` | Eventos del ciclo de vida de reservas. |
| `origenx.inventory.events` | Eventos de stock e inventario. |
| `origenx.notifications.events` | Eventos de solicitud y resultado de notificaciones. |
| `origenx.gateway.events` | Eventos emitidos desde el API Gateway para trazabilidad de acciones externas. |
| `origenx.events.dlq` | Dead letter queue para eventos que no pudieron procesarse correctamente. |

## Eventos definidos

| Evento | Tópico | Productor | Consumidores | Proposito |
|---|---|---|---|---|
| `UserCreated` | `origenx.users.events` | `user-service` | `notification-service`, `api-gateway`, auditoria futura | Informar que un usuario fue creado. |
| `UserUpdated` | `origenx.users.events` | `user-service` | `api-gateway`, auditoria futura | Informar cambios en el perfil de un usuario. |
| `ReservationCreated` | `origenx.reservations.events` | `reservation-service` | `notification-service`, auditoria futura | Registrar que una solicitud de reserva fue creada. |
| `ReservationConfirmed` | `origenx.reservations.events` | `reservation-service` | `notification-service`, `api-gateway`, auditoria futura | Informar que una reserva fue confirmada. |
| `ReservationFailed` | `origenx.reservations.events` | `reservation-service` | `notification-service`, `api-gateway`, auditoria futura | Informar que una reserva fallo. |
| `ReservationCancelled` | `origenx.reservations.events` | `reservation-service` | `notification-service`, `inventario-service`, auditoria futura | Informar que una reserva fue cancelada. |
| `InventoryStockBlocked` | `origenx.inventory.events` | `inventario-service` | `reservation-service`, auditoria futura | Informar que se bloqueo stock para una habitacion. |
| `InventoryStockReleased` | `origenx.inventory.events` | `inventario-service` | `reservation-service`, auditoria futura | Informar que se libero stock previamente bloqueado. |
| `InventoryStockFailed` | `origenx.inventory.events` | `inventario-service` | `reservation-service`, auditoria futura | Informar que una operacion de stock fallo. |
| `NotificationRequested` | `origenx.notifications.events` | `reservation-service` | `notification-service`, auditoria futura | Solicitar el envio asincronico de una notificacion. |
| `NotificationSent` | `origenx.notifications.events` | `notification-service` | `reservation-service`, auditoria futura | Informar que una notificacion fue enviada correctamente. |
| `NotificationFailed` | `origenx.notifications.events` | `notification-service` | `reservation-service`, auditoria futura | Informar que una notificacion fallo. |
| `GatewayUserRegistered` | `origenx.gateway.events` | `api-gateway` | auditoria futura, analitica futura | Registrar una solicitud externa de registro de usuario. |
| `GatewayUserLoggedIn` | `origenx.gateway.events` | `api-gateway` | auditoria futura, analitica futura | Registrar un inicio de sesion exitoso desde el gateway. |
| `GatewayInventorySearchRequested` | `origenx.gateway.events` | `api-gateway` | auditoria futura, analitica futura | Registrar una busqueda de inventario iniciada por un cliente. |
| `GatewayReservationRequested` | `origenx.gateway.events` | `api-gateway` | auditoria futura, analitica futura | Registrar una solicitud externa de reserva. |

> Nota: los consumidores indicados son consumidores previstos. En esta fase aun no hay consumidores Kafka implementados.

## Reglas de versionado

- Cada evento debe incluir `version`.
- La primera version estable de cada contrato es `1`.
- Los cambios compatibles deben mantener la misma version:
  - agregar campos opcionales al `payload`;
  - agregar valores nuevos que los consumidores puedan ignorar;
  - ampliar metadata sin cambiar significado existente.
- Los cambios incompatibles deben crear una nueva version:
  - renombrar campos existentes;
  - eliminar campos usados por consumidores;
  - cambiar tipo o significado de un campo;
  - cambiar reglas obligatorias del `payload`.
- Los consumidores deben ignorar campos desconocidos para facilitar evolucion gradual.

## Reglas de idempotencia

- `event_id` es obligatorio y unico.
- Un consumidor debe poder procesar el mismo evento mas de una vez sin duplicar efectos.
- Para eventos de negocio puede usarse una clave natural adicional:
  - `user_id` para eventos de usuario;
  - `reservation_id` para eventos de reserva;
  - `room_type_id` y `operation_id` para eventos de inventario;
  - `notification_id` o `reservation_id` + `notification_type` para notificaciones.
- Si un evento se reintenta, debe conservar el mismo `event_id` cuando representa el mismo hecho.

## Reglas de trazabilidad

- `occurred_at` debe estar en UTC y formato ISO-8601.
- `correlation_id` debe propagarse cuando una accion cruza varios servicios.
- `source_service` debe usar el nombre operativo del servicio:
  - `api-gateway`
  - `user-service`
  - `inventario-service`
  - `reservation-service`
  - `notification-service`
- Los logs de productores y consumidores deben incluir `event_id` y `correlation_id` cuando exista.
- Cuando un evento falle despues de reintentos, debe enviarse a `origenx.events.dlq` con informacion suficiente para diagnostico.

## Dead letter queue

El topico `origenx.events.dlq` debe usarse para eventos que no pudieron ser procesados tras los reintentos definidos.

El `payload` de un evento DLQ debe incluir al menos:

- evento original;
- topico original;
- nombre del consumidor que fallo;
- motivo del error;
- fecha del fallo;
- cantidad de intentos.

No se debe incluir informacion sensible adicional en la DLQ.

## Datos prohibidos en eventos

No deben publicarse en Kafka:

- passwords;
- hashes de passwords;
- access tokens;
- refresh tokens;
- API keys;
- secretos de servicios;
- credenciales de bases de datos;
- datos completos de tarjetas de pago;
- documentos de identidad u otros datos sensibles no requeridos;
- payloads HTTP completos si contienen headers de autenticacion.

Si un dato personal es necesario para un caso de negocio, debe publicarse minimizado. Por ejemplo, preferir `user_id` sobre datos completos del usuario.

## Ejemplos

Los ejemplos JSON estan en:

```text
contracts/events/examples/
```

Cada archivo representa un evento valido con el sobre comun y un `payload` especifico.
