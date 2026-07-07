# Guion de video demo: Integración Kafka en Origen X

## 1. Objetivo del video

El video debe mostrar la integración de Kafka en Origen X como bus transversal de eventos. La demo debe dejar claro que gRPC se mantiene para la comunicación interna principal, que las acciones reales del frontend generan eventos y que esos eventos se pueden validar desde Kafka UI o consola.

El flujo recomendado incluye registro de usuario, login, búsqueda de habitaciones y creación de reserva.

## 2. Duración objetivo

El video debe durar máximo 3 minutos.

| Tiempo | Escena | Objetivo |
|---|---|---|
| 0:00 - 0:20 | Presentación | Explicar qué es Origen X |
| 0:20 - 0:45 | Arquitectura | REST, gRPC y Kafka |
| 0:45 - 1:10 | Implementación | Mostrar servicios/tópicos |
| 1:10 - 2:25 | Demo funcional | Crear usuario, login, búsqueda, reserva |
| 2:25 - 2:45 | Evidencia Kafka | Mostrar eventos |
| 2:45 - 3:00 | Cierre | Conclusión técnica |

## 3. Preparación antes de grabar

Checklist:

- Docker Desktop abierto.
- Proyecto abierto en VS Code o editor.
- Servicios levantados.
- Frontend abierto.
- Kafka UI abierto.
- Terminal preparada.
- Navegador con `http://localhost:3000`.
- Navegador con `http://localhost:8082`.
- Si se usarán consumers, abrir solo los necesarios para no saturar el video.

## 4. Comandos previos

Levantar servicios:

```bash
docker compose up -d
docker compose ps
```

Consumer recomendado para mostrar reservas:

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.reservations.events --from-beginning'
```

Consumer recomendado para mostrar gateway:

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.gateway.events --from-beginning'
```

Consumer recomendado para mostrar notificaciones:

```bash
docker compose exec kafka sh -lc '/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic origenx.notifications.events --from-beginning'
```

Para un video de 3 minutos, es mejor usar Kafka UI o máximo 2 terminales. No conviene abrir cinco consumers a la vez porque se pierde claridad visual.

## 5. Estructura del video en 3 minutos

### Escena 1: Presentación del proyecto

Duración: 20 segundos.

Qué mostrar:

- Frontend o README.

Qué decir:

"Origen X es una plataforma distribuida de reservas hoteleras basada en microservicios. El sistema permite registrar usuarios, iniciar sesión, buscar habitaciones, crear reservas y generar notificaciones."

### Escena 2: Arquitectura y decisión técnica

Duración: 25 segundos.

Qué mostrar:

- Diagrama o README.

Qué decir:

"La arquitectura mantiene REST desde el frontend hacia el API Gateway y gRPC para la comunicación interna entre microservicios. Kafka fue agregado como bus transversal de eventos, no como reemplazo de gRPC."

### Escena 3: Implementación Kafka

Duración: 25 segundos.

Qué mostrar:

- Kafka UI o README con tabla de tópicos.

Qué decir:

"Se implementaron tópicos por dominio: gateway, usuarios, reservas, inventario y notificaciones. Cada servicio publica eventos relevantes de su operación."

Mencionar:

- api-gateway publica eventos de auditoría.
- user-service publica eventos de usuario.
- reservation-service publica eventos de reserva.
- inventario-service publica eventos de stock.
- notification-service consume eventos de reserva y publica eventos de notificación.

### Escena 4: Demo desde frontend

Duración: 1 minuto 15 segundos.

Qué hacer:

1. Crear usuario o usar uno nuevo.
2. Iniciar sesión.
3. Buscar habitaciones.
4. Crear una reserva.

Qué decir:

"Ahora vamos a ejecutar acciones reales desde el frontend. Al registrar, iniciar sesión, buscar y reservar, los microservicios siguen respondiendo mediante REST y gRPC, mientras Kafka registra los eventos generados."

Eventos esperados:

- GatewayUserRegistered.
- UserCreated.
- GatewayUserLoggedIn.
- UserLoggedIn.
- GatewayInventorySearchRequested.
- GatewayReservationRequested.
- ReservationCreated.
- InventoryStockBlocked.
- ReservationConfirmed.
- NotificationSent.

### Escena 5: Evidencia en Kafka

Duración: 20 segundos.

Qué mostrar:

- Kafka UI o consumer.

Qué decir:

"Aquí se observan los eventos publicados en Kafka. Por ejemplo, al crear una reserva aparecen eventos de gateway, reserva, inventario y notificación."

### Escena 6: Idempotencia y tolerancia a fallos

Duración: 15 segundos.

Qué decir:

"En notification-service se agregó idempotencia usando reservation-confirmation:<reservation_id>, evitando duplicar notificaciones si la confirmación llega por gRPC y también por Kafka."

### Escena 7: Cierre

Duración: 15 segundos.

Qué decir:

"Con esto se demuestra una integración distribuida donde gRPC mantiene las operaciones sincrónicas y Kafka agrega eventos asincrónicos para trazabilidad, auditoría y desacoplamiento."

## 6. Guion hablado completo de 3 minutos

"En esta demo mostramos la integración de Kafka en Origen X, una plataforma distribuida de reservas hoteleras basada en microservicios.

El sistema mantiene una arquitectura donde el frontend se comunica por REST con el API Gateway, y el API Gateway coordina los microservicios internos usando gRPC. Kafka no reemplaza gRPC. gRPC se mantiene para operaciones que requieren respuesta inmediata, como validar usuarios, consultar inventario, bloquear stock y crear reservas.

Kafka se usa como bus de eventos transversal. Esto permite publicar eventos asincrónicos para auditoría, trazabilidad y desacoplamiento, sin cambiar los endpoints ni los contratos Protobuf existentes.

En la implementación se definieron tópicos por dominio. El API Gateway publica eventos de auditoría en origenx.gateway.events. user-service publica eventos de usuario en origenx.users.events. reservation-service publica eventos de reserva en origenx.reservations.events. inventario-service publica eventos de stock en origenx.inventory.events. Y notification-service consume ReservationConfirmed desde Kafka y publica NotificationSent o NotificationFailed en origenx.notifications.events.

Ahora ejecutamos acciones reales desde el frontend. Primero registramos un usuario, luego iniciamos sesión, buscamos habitaciones disponibles y finalmente creamos una reserva. La reserva sigue funcionando desde el frontend usando REST hacia el API Gateway y gRPC entre microservicios.

Al crear una reserva se publican eventos en distintos tópicos: GatewayReservationRequested, ReservationCreated, InventoryStockBlocked y ReservationConfirmed. Luego notification-service consume ReservationConfirmed y publica NotificationSent.

También se implementó idempotencia en notification-service usando la clave reservation-confirmation:<reservation_id>. Esto evita duplicar notificaciones si la confirmación llega por gRPC y también por Kafka.

La integración es best effort, por lo tanto Kafka no rompe el flujo principal si falla. Si Kafka no está disponible, los servicios registran el error, pero la operación principal puede continuar por REST y gRPC.

Con esto se demuestra que Origen X mantiene gRPC para la coordinación sincrónica entre microservicios y agrega Kafka como una capa transversal de eventos asincrónicos para mejorar trazabilidad, auditoría y capacidad de extensión."

## 7. Qué mostrar en pantalla

1. Frontend en `http://localhost:3000`.
2. Kafka UI en `http://localhost:8082`.
3. `docker compose ps` mostrando Kafka healthy.
4. Tabla de tópicos o Kafka UI.
5. Acción de crear, buscar y reservar.
6. Eventos generados.

## 8. Checklist de eventos esperados

| Acción | Tópico | Evento esperado |
|---|---|---|
| Registro | `origenx.gateway.events` | GatewayUserRegistered |
| Registro | `origenx.users.events` | UserCreated |
| Login | `origenx.gateway.events` | GatewayUserLoggedIn |
| Login | `origenx.users.events` | UserLoggedIn |
| Búsqueda | `origenx.gateway.events` | GatewayInventorySearchRequested |
| Reserva | `origenx.gateway.events` | GatewayReservationRequested |
| Reserva | `origenx.reservations.events` | ReservationCreated |
| Reserva | `origenx.inventory.events` | InventoryStockBlocked |
| Reserva | `origenx.reservations.events` | ReservationConfirmed |
| Reserva | `origenx.notifications.events` | NotificationSent |
| Actualización de perfil | `origenx.users.events` | UserUpdated |

## 9. Problemas comunes durante la grabación

- Puerto 8080 ocupado: detener el proceso que usa el puerto o cambiar el mapeo temporalmente.
- `localhost:3000` no carga: revisar que el contenedor `frontend` esté arriba.
- Git Bash transforma `/opt/kafka`: usar `docker compose exec kafka sh -lc '...'`.
- Tópico aparece vacío: generar una acción nueva desde el frontend.
- Kafka todavía no está healthy: esperar y revisar `docker compose ps`.
- Warnings `RESEND_API_KEY` y `USER_SERVICE_HOST`: no bloquean la demo de Kafka.
- notification-service detecta duplicado: es esperado si la confirmación llega por gRPC y Kafka.
- Consumer no muestra eventos antiguos: puede depender del tópico, grupo o de que no existan eventos previos.

## 10. Plan B para video de 3 minutos

- Usar Kafka UI en vez de muchas terminales.
- Mostrar solo `origenx.gateway.events` y `origenx.reservations.events`.
- Mostrar capturas previas si no se alcanzan a ver todos los eventos.
- Crear un usuario nuevo para forzar eventos nuevos.
- Usar una reserva nueva para forzar ReservationCreated y ReservationConfirmed.
- Tener los servicios ya levantados antes de iniciar grabación.

## 11. Cierre recomendado

"Con esta implementación, Origen X mantiene gRPC para la coordinación sincrónica entre microservicios y agrega Kafka como una capa transversal de eventos asincrónicos, mejorando la trazabilidad, auditoría y capacidad de extensión del sistema."
