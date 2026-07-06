# Enzo Loren
# Documento individual del bloque - Resiliencia

## 1. Descripcion del bloque

El bloque de resiliencia cubre como el sistema tolera fallos parciales sin caerse por completo ni degradar la experiencia de forma silenciosa. Se implemento sobre dos puntos del sistema:

- `api-gateway`: unico punto de entrada HTTP, donde viven el rate limiting, los timeouts hacia los servicios internos, el retry con backoff exponencial y el fallback de busqueda de inventario.
- `reservation_service` (`mi-servicio`): orquestador de la SAGA de reservas, donde viven los circuit breakers hacia `user-service`, `inventario_service` y `notification_service`, y la idempotencia de `CreateReservation`.

No se toco `user-service`, `inventario_service` ni `notification_service` por dentro: el circuit breaker vive en quien llama (`reservation_service`), no en quien responde, y el rate limiting vive en el unico punto de entrada (`api-gateway`), no repartido en cada microservicio.

El objetivo del bloque es que una falla en un servicio (caida, lentitud, sobrecarga) no se propague como una falla identica o peor en los servicios que dependen de el, y que los clientes reciban una respuesta rapida y con un codigo de error que explique de quien es el problema (suyo por exceso de trafico, o del sistema por una dependencia caida) en vez de quedarse esperando o recibir un 500 generico.

## 2. Decisiones tecnicas

### Decision 1: circuit breaker propio en vez de libreria externa

Se implemento un circuit breaker propio (`mi-servicio/circuitbreaker.go`) con tres estados (`CLOSED`, `OPEN`, `HALF_OPEN`), en vez de usar una libreria como `sony/gobreaker`.

Alternativa descartada: usar `gobreaker`.

Razonamiento: `reservation_service` ya lleva su propio manejo de estado a mano para la SAGA (tablas `reservation_sagas` y `reservation_saga_events`, transiciones registradas explicitamente en `recordSagaTransition`). Meter una libreria externa solo para el circuit breaker habria mezclado dos formas distintas de manejar estado en el mismo servicio: una hecha a mano y visible en metricas propias, y otra opaca dentro de una libreria. Implementarlo a mano cuesta menos de 100 lineas, deja las transiciones de estado visibles como metricas Prometheus (`circuit_breaker_state`, `circuit_breaker_transitions_total`) igual que las metricas de la SAGA, y permite explicar exactamente que dispara cada cambio de estado en el video sin depender del comportamiento interno de un paquete de terceros.

Se instancio un breaker independiente por dependencia (`userBreaker`, `inventoryBreaker`, `notificationBreaker`) en vez de un unico breaker global para todas las llamadas salientes. Esto es especifico al dominio: si `notification_service` esta lento o caido, no tiene sentido que las llamadas a `user-service` o `inventario_service` tambien empiecen a fallar rapido, porque son dependencias independientes con su propia disponibilidad.

### Decision 2: dos algoritmos de rate limiting distintos segun el endpoint

Se implemento un token bucket global (`api-gateway/rate_limiter.py`, clase `TokenBucketLimiter`) aplicado a todas las rutas HTTP, con capacidad de 20 tokens y recarga de 10 tokens por segundo por cliente (identificado por `user_id` si hay sesion, o por IP si no la hay). Ademas, se implemento un sliding window log (`SlidingWindowLimiter`) aplicado unicamente a `POST /login`, con un limite de 5 intentos por 60 segundos por combinacion de email e IP.

Alternativa descartada: aplicar el mismo sliding window estricto a todos los endpoints.

Razonamiento: el mismo bloque agrega retry con backoff para operaciones idempotentes (busqueda, listados, creacion de reserva). Si se aplicara un sliding window estricto de forma global, los reintentos automaticos que genera el propio gateway ante un `UNAVAILABLE` transitorio contarian como trafico abusivo del cliente y podrian terminar bloqueando al mismo cliente que el sistema esta tratando de ayudar con el retry. El token bucket permite absorber esas rafagas cortas y legitimas (un reintento, un doble click) sin penalizar al cliente, mientras sigue limitando el promedio sostenido. En cambio, en `/login` una rafaga de intentos no es trafico legitimo retransmitido por el propio sistema: es el patron de un ataque de fuerza bruta o credential stuffing, asi que ahi si conviene un limite estricto que no perdone rafagas.

La respuesta ante limite excedido es siempre HTTP 429 con encabezado `Retry-After`, calculado a partir del tiempo de recarga del bucket o del tiempo restante de la ventana segun corresponda. Este 429 se diferencia deliberadamente del 503 que devuelve el circuit breaker: el 429 dice "estas pidiendo demasiado rapido" (responsabilidad del cliente), el 503 dice "una dependencia no esta disponible" (responsabilidad del sistema).

### Decision 3: idempotencia persistida en la base de reservas, no en memoria del gateway

`CreateReservation` ahora recibe un campo `idempotency_key` (agregado a `CreateReservationRequest` en el `.proto` compartido). El `api-gateway` genera una clave (`uuid4`) una sola vez por peticion HTTP entrante si el cliente no envio su propio header `Idempotency-Key`, y reutiliza esa misma clave en todos los reintentos internos de esa peticion. `reservation_service` persiste esa clave junto con la reserva en `reservas_db` (columna `idempotency_key`, con restriccion unica), y antes de iniciar una SAGA nueva revisa si ya existe una reserva con esa clave; si existe, devuelve la reserva ya creada en vez de volver a bloquear stock y crear una reserva duplicada.

Alternativa descartada: deduplicar en memoria dentro del propio `api-gateway` (un diccionario de claves ya vistas).

Razonamiento: aunque hoy `api-gateway` corre como un solo proceso, deduplicar en memoria ata la garantia de idempotencia a la vida del proceso: un reinicio del contenedor del gateway borra el registro de claves usadas, y una respuesta que se perdio justo antes del reinicio podria reintentarse despues y crear una reserva duplicada. Persistir la clave en `reservas_db`, la misma base que ya es la fuente de verdad de las reservas, ata la garantia de idempotencia al mismo dato que nunca deberia duplicarse, sin depender de que el gateway seguido con vida recuerde algo.

## 3. Comportamiento ante fallos

**`user-service` cae o deja de responder:** las primeras 5 llamadas consecutivas de `reservation_service` hacia `GetUser` fallan por timeout o error de transporte. Al quinto fallo consecutivo, `userBreaker` pasa a `OPEN`. Mientras esta abierto, `CreateReservation` ni siquiera intenta la llamada gRPC: falla de inmediato con `codes.Unavailable`, se registra `sagaUserValidationFailed` y `sagaFailed` igual que un fallo real de validacion, y se incrementa `circuit_breaker_short_circuits_total{dependency="user"}`. Pasados 15 segundos, el breaker pasa a `HALF_OPEN` y deja pasar una sola llamada de prueba; si responde bien, vuelve a `CLOSED`; si falla, vuelve a `OPEN` y reinicia el temporizador.

**`inventario_service` cae o deja de responder:** mismo patron que el anterior pero con `inventoryBreaker`, de forma independiente. Una caida de inventario no afecta el estado del breaker de usuarios ni el de notificaciones. Esto evita que una sola dependencia caida genere fallos en cascada sobre llamadas que no dependen de ella.

**`notification_service` cae o deja de responder:** al abrirse `notificationBreaker`, el goroutine que dispara la notificacion no intenta la llamada gRPC; directamente registra `sagaCompletedWithNotificationFailure`, el mismo estado que la SAGA ya usaba para un fallo real de notificacion. La reserva sigue confirmada porque la notificacion nunca fue parte del camino critico. La diferencia frente a un fallo puntual es que, con el breaker abierto, ni siquiera se gasta un timeout de 3 segundos por reserva intentando una llamada que ya se sabe que va a fallar.

**Un cliente excede el limite de trafico:** al superar 20 peticiones de rafaga o sostener mas de 10 por segundo, el `api-gateway` responde 429 con `Retry-After` sin reenviar la peticion a ningun servicio interno. Un intento de fuerza bruta contra `/login` (mas de 5 intentos en 60 segundos desde la misma IP y email) recibe 429 aunque el limite global de token bucket todavia tenga cupo disponible.

**`reservation_service` no responde dentro del timeout:** el `api-gateway` corta la espera de `CreateReservation` a los 12 segundos (tiempo mayor a la suma de los timeouts internos de la SAGA hacia usuario e inventario, 5s + 5s, para no cortar una operacion que todavia puede completarse del lado del servidor). Si el timeout se cumple, el gateway responde con error de dependencia en vez de dejar la conexion HTTP colgada indefinidamente.

**Un error transitorio de transporte en una operacion de lectura:** una llamada a `ListReservations` o `SearchAvailableRooms` que falla con `UNAVAILABLE` o `DEADLINE_EXCEEDED` se reintenta automaticamente hasta 2 veces mas, con backoff exponencial (250ms, 500ms) mas un jitter aleatorio de hasta 100ms. Un error de negocio (`NOT_FOUND`, `UNAUTHENTICATED`, `RESOURCE_EXHAUSTED`) nunca se reintenta: reintentar "no hay stock" o "credenciales invalidas" no cambia el resultado y solo agrega latencia.

**Una respuesta se pierde despues de que la reserva ya se confirmo:** si `reservation_service` confirma la reserva pero la respuesta no llega al gateway por un corte de red, y el gateway reintenta `CreateReservation` con la misma `idempotency_key`, `reservation_service` detecta la clave ya usada y devuelve la reserva existente en vez de bloquear stock una segunda vez y crear una reserva duplicada.

**`inventario_service` no responde durante una busqueda:** despues de agotar los reintentos, `POST /api/inventory/search` no responde 500. Responde 200 con `{"rooms": [], "degraded": true}`, permitiendo que el frontend muestre "no se pudo cargar disponibilidad en este momento" en vez de un error generico.

## 4. Trade-offs y limitaciones

El estado de cada circuit breaker vive en memoria del proceso de `reservation_service`. Con una sola instancia del servicio esto es correcto, pero si el servicio corriera con mas de una replica, cada una llevaria su propio estado de breaker de forma independiente: una replica podria tener el breaker de inventario abierto mientras otra lo tiene cerrado, porque no hay un estado compartido entre procesos.

Los umbrales del circuit breaker (5 fallos consecutivos, 15 segundos de apertura, 1 llamada de prueba en half-open) se fijaron por observacion practica probando el sistema, no a partir de un SLO medido con trafico real. En un entorno productivo estos numeros se calibrarian con datos historicos de latencia y tasa de error por dependencia.

El rate limiter tambien vive en memoria del proceso de `api-gateway`: un reinicio del contenedor hace que todos los clientes recuperen su cupo completo de inmediato. Con una sola instancia de gateway esto no genera inconsistencia entre replicas, pero tampoco persiste el historial de consumo entre reinicios.

El retry con backoff solo cubre errores de transporte (`UNAVAILABLE`, `DEADLINE_EXCEEDED`). Un fallo que en realidad es transitorio pero que el servicio downstream reporta con un codigo de error de negocio no se beneficia del retry, porque el gateway no puede distinguir "negocio real" de "negocio por una falla interna momentanea" solo mirando el codigo gRPC.

La idempotencia cubre especificamente `CreateReservation`, que es la unica operacion de escritura del bloque que crea un recurso nuevo con efectos secundarios costosos de deshacer (bloqueo de stock, notificacion). Otras escrituras como `UpdateUser` no tienen clave de idempotencia propia, lo cual se acepta porque ya son naturalmente idempotentes (sobrescriben campos en vez de incrementarlos), pero es una decision que depende de que esa propiedad se mantenga si esas operaciones cambian en el futuro.

El fallback de busqueda degrada a una lista vacia con `degraded: true`, no a un resultado cacheado. Esto es mas simple de razonar y de demostrar, pero significa que el usuario ve "sin resultados" en vez de una disponibilidad ligeramente desactualizada pero util.

## 5. Ubicacion en el codigo

`api-gateway`:

- `rate_limiter.py`: `TokenBucketLimiter` (conectado como `RateLimitMiddleware`, aplicado a toda ruta HTTP excepto `/health`) y `SlidingWindowLimiter` (invocado directamente desde el endpoint `/login`, ya con el body parseado, para no tener que leer el stream de la peticion dos veces).
- `resilience.py`: helper `call_with_retry` (backoff exponencial + jitter, filtrado por codigo gRPC), constantes de timeout por tipo de llamada, generacion de `idempotency_key` por peticion.
- `main.py`: wiring del middleware de rate limiting, uso de `call_with_retry` en los endpoints, manejo del header `Idempotency-Key`, fallback de `/api/inventory/search`.
- `reservations_client.py`, `users_client.py`, `inventory_client.py`: cada llamada gRPC recibe un timeout explicito.

`mi-servicio` (reservation_service):

- `circuitbreaker.go`: struct `CircuitBreaker`, estados `CLOSED`/`OPEN`/`HALF_OPEN`, metricas asociadas.
- `main.go`: campos `userBreaker`, `inventoryBreaker`, `notificationBreaker` en `reservationServer`; las tres llamadas salientes (`GetUser`, `UpdateStock`, `SendConfirmation`) pasan por su breaker correspondiente antes de intentarse; columna `idempotency_key` agregada en `initializeReservationSchema`; verificacion de idempotencia al inicio de `CreateReservation`.
- `pb/servicio.proto`: campo `idempotency_key` agregado a `CreateReservationRequest` (sincronizado con la copia en `api-gateway/proto/servicio.proto`).

## 6. Metricas expuestas

En `reservation_service` (scrapeadas por Prometheus en el job existente `reservation-service`):

- `circuit_breaker_state{dependency}`: 0 = closed, 1 = open, 2 = half_open.
- `circuit_breaker_transitions_total{dependency,to_state}`.
- `circuit_breaker_short_circuits_total{dependency}`: llamadas rechazadas sin intentar el RPC.
- `reservations_idempotent_replays_total`: veces que se devolvio una reserva existente en vez de crear una nueva.

En `api-gateway` no se agregaron metricas Prometheus nuevas (el gateway hoy no expone `/metrics` y no esta en el scrape config de `observability/prometheus.yml`); el comportamiento de rate limiting, retry y fallback queda visible mediante logs estructurados existentes (`logging.warning`/`logging.info`) siguiendo el mismo estilo que ya usa `main.py`.

## 7. Donde verlo funcionando

Circuit breaker de notificaciones:

```bash
docker compose stop notification_service
# crear una reserva vía POST /reservations con Authorization: Bearer <TOKEN>
# (el <TOKEN> se obtiene del access_token que devuelve POST /login)
docker compose logs --tail 30 reservation_service
# esperar transición a OPEN tras 5 fallos, luego:
docker compose start notification_service
# la siguiente reserva dispara el intento HALF_OPEN y cierra el breaker si responde bien
```

Rate limiting en login:

```bash
for i in $(seq 1 6); do
  curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/login \
    -H "Content-Type: application/json" \
    -d '{"email":"a@test.com","password":"mala"}'
done
# los primeros 5 responden 401, el sexto responde 429 con Retry-After
```

Idempotencia:

```bash
curl -X POST http://localhost:8080/reservations \
  -H "Content-Type: application/json" -H "Authorization: Bearer <TOKEN>" -H "Idempotency-Key: demo-123" \
  -d '{"hotel_id":"h1","room_type_id":"rt2","fecha_inicio":"2026-06-01","fecha_fin":"2026-06-10"}'
# repetir la misma llamada con el mismo header: debe devolver el mismo reservation_id
```

Fallback de busqueda:

```bash
docker compose stop inventario_service
curl -X POST http://localhost:8080/api/inventory/search \
  -H "Content-Type: application/json" \
  -d '{"fecha_inicio":"2026-06-01","fecha_fin":"2026-06-10"}'
# responde 200 con {"rooms": [], "degraded": true} en vez de 500
```
