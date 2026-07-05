# Bloque SAGA - Reservation Service

## Descripcion del bloque

El bloque SAGA se implementa en el flujo de creacion de reservas. Este flujo involucra cuatro servicios del sistema:

- `api-gateway`: recibe la solicitud HTTP `POST /reservations` y valida el token del usuario.
- `reservation_service`: orquesta la SAGA y decide que paso ejecutar o compensar.
- `user-service`: valida que el usuario exista antes de crear la reserva.
- `inventario_service`: bloquea o libera stock de la habitacion.
- `notification_service`: registra y envia la notificacion de confirmacion.

El problema distribuido es que no existe una transaccion unica entre todas las bases de datos. Usuarios, inventario, reservas y notificaciones tienen persistencia separada y se comunican por gRPC. Por eso la reserva se maneja como una SAGA: una secuencia de pasos locales con compensaciones cuando un paso posterior falla.

Flujo implementado:

1. `STARTED`: se crea una instancia de SAGA con `saga_id`.
2. `USER_VALIDATED`: se valida el usuario en `user-service`.
3. `STOCK_LOCKED`: se bloquea una unidad de stock en `inventario_service`.
4. `RESERVATION_CREATED`: se guarda la reserva en `reservas_db`.
5. `NOTIFICATION_DISPATCHING`: se dispara la notificacion asincrona.
6. `COMPLETED`: la reserva queda confirmada y la notificacion fue procesada.

Cada transicion queda registrada en dos lugares: logs del contenedor `reservation_service` y tablas `reservation_sagas` / `reservation_saga_events` en `reservas_db`.

## Decision tecnica principal

Se eligio una SAGA orquestada desde `reservation_service`.

Alternativa descartada: SAGA coreografiada basada en eventos entre servicios.

Razonamiento: el flujo de reserva ya estaba centralizado en `reservation_service`, que llama a usuario, inventario y notificaciones. Usar una SAGA orquestada permite mantener la implementacion simple, demostrable y coherente con el codigo existente. Tambien facilita explicar el flujo en el video, porque hay un punto unico donde se ve cada paso, cada fallo y cada compensacion.

Otra decision fue persistir el estado en PostgreSQL, no solo imprimir logs. Los logs sirven para observabilidad en tiempo real, pero las tablas permiten auditar que paso con una reserva incluso despues de reiniciar contenedores.

## Comportamiento ante fallos

Fallo validando usuario:

- Estado registrado: `USER_VALIDATION_FAILED` y luego `FAILED`.
- No hay compensacion porque todavia no se bloqueo stock.
- El cliente recibe error gRPC `Unauthenticated` desde reservas y el gateway lo transforma en error HTTP.

Fallo bloqueando inventario:

- Estado registrado: `STOCK_LOCK_FAILED` o `STOCK_UNAVAILABLE`, y luego `FAILED`.
- No hay compensacion porque no se desconto stock.
- Se incrementa la metrica `reservations_failures_total{stage="inventory_lock"}` o `inventory_no_stock`.

Fallo guardando la reserva despues de bloquear stock:

- Estado registrado: `RESERVATION_INSERT_FAILED`.
- La SAGA entra a `COMPENSATING`.
- Se llama a `inventario_service.UpdateStock` con accion `LIBERAR`.
- La compensacion se reintenta hasta 3 veces.
- Si libera stock, queda evento `COMPENSATED` y `compensation_status=DONE`.
- Si no puede liberar stock, queda evento `COMPENSATION_FAILED` y `compensation_status=FAILED`.

Fallo enviando notificacion:

- La reserva no se cancela, porque la reserva ya fue confirmada y el stock ya fue bloqueado correctamente.
- Se registra `COMPLETED_WITH_NOTIFICATION_FAILURE`.
- El sistema degrada parcialmente: la reserva existe, pero la notificacion puede revisarse o reintentarse despues.

Fallo persistiendo el inicio de la SAGA:

- No se ejecuta ningun llamado externo.
- El flujo falla antes de tocar usuario, inventario o notificaciones.
- Esto evita efectos secundarios sin auditoria.

## Trade-offs y limitaciones

La SAGA implementada no usa Kafka ni una cola de mensajes. Esto se acepto porque el bloque asignado es SAGA y el sistema actual ya usa gRPC sincrono entre servicios. Para esta entrega, una SAGA orquestada es mas directa y facil de demostrar.

La compensacion de inventario se ejecuta desde `reservation_service` con tres reintentos simples. Si `inventario_service` esta caido por mucho tiempo, la compensacion queda marcada como `COMPENSATION_FAILED`. La ventaja es que el fallo queda visible y auditable; la limitacion es que no hay un worker posterior que reintente indefinidamente.

La notificacion se trata como paso no critico. Si falla, no se cancela la reserva. Este trade-off evita anular reservas validas solo porque fallo un email, pero implica que el sistema necesita un mecanismo futuro de reintento de notificaciones.

El monto de la reserva sigue siendo fijo (`150.50`) porque el objetivo del bloque es consistencia distribuida y compensacion, no calculo tarifario.

## Ubicacion en el codigo

Archivo principal:

- `mi-servicio/main.go`

Funciones relevantes:

- `CreateReservation`: orquesta la SAGA completa.
- `startSaga`: crea la instancia inicial en `reservation_sagas`.
- `recordSagaTransition`: registra transiciones en logs, metricas y `reservation_saga_events`.
- `failSaga`: marca fallos finales.
- `compensateInventory`: libera stock cuando falla un paso posterior al bloqueo.
- `initializeReservationSchema`: crea/actualiza las tablas de SAGA al iniciar el servicio.

## Donde verlo funcionando

Logs en tiempo real:

```bash
docker compose logs -f reservation_service
```

Filtrar en Grafana/Loki:

```text
{service="reservation_service"} |= "saga_id"
```

Ver eventos persistidos:

```bash
docker exec -it reservas_db psql -U postgres -d reservas_service -c "SELECT saga_id, status, current_step, compensation_status, error_message FROM reservation_sagas ORDER BY created_at DESC LIMIT 5;"
```

```bash
docker exec -it reservas_db psql -U postgres -d reservas_service -c "SELECT saga_id, step, result, detail, created_at FROM reservation_saga_events ORDER BY created_at DESC LIMIT 20;"
```

Metricas Prometheus:

```text
reservation_saga_transitions_total
reservation_saga_compensations_total
reservations_failures_total
```

## Guion breve para video de 3 minutos

Este bloque implementa el patron SAGA en la creacion de reservas. En un sistema distribuido no podemos usar una transaccion normal, porque usuarios, inventario, reservas y notificaciones tienen bases separadas y se comunican por gRPC. El riesgo era dejar el sistema inconsistente, por ejemplo bloquear stock pero fallar al guardar la reserva.

La SAGA esta implementada en `reservation_service`, especificamente en `CreateReservation`. Este servicio actua como orquestador: genera un `saga_id`, valida el usuario, bloquea stock, guarda la reserva y dispara la notificacion. Cada paso queda registrado en logs, metricas y tablas de auditoria.

La decision tecnica fue usar SAGA orquestada en vez de coreografiada. La descarte coreografiada porque el proyecto ya tenia este flujo centralizado en reservas, y para esta entrega era mas claro mantener un punto unico de control y compensacion.

Ante fallos, la SAGA decide que hacer. Si falla el usuario o inventario, termina como `FAILED`. Si falla la base despues de bloquear stock, ejecuta compensacion llamando a inventario con `LIBERAR`. Si falla notificacion, no cancela la reserva: queda `COMPLETED_WITH_NOTIFICATION_FAILURE`, porque un email fallido no deberia anular una reserva confirmada.

La ventaja es que ahora se puede demostrar consistencia eventual, compensaciones y observabilidad distribuida. En los logs se busca `saga_id`, y en la base se revisan `reservation_sagas` y `reservation_saga_events` para ver todo el recorrido.
