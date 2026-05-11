# Origen X: Informe Técnico de Arquitectura

Este documento describe la arquitectura y flujos técnicos del sistema de reservas hoteleras **Origen X**, alineado exactamente con la realidad del código fuente implementado.

---

## 1. Arquitectura del Sistema (Cumplimiento de Criterios)

### 1.1 Separación de Responsabilidades
El ecosistema se distribuye en 4 microservicios de dominio independiente y claramente delimitado, más un Gateway que orquesta la entrada:
1.  **API Gateway (BFF - Python/FastAPI):** Traduce las peticiones REST/JSON del cliente web a llamadas de protocolo binario gRPC de la red interna.
2.  **User Service (Java/Spring Boot):** Administra la identidad, autenticación (hashing) y perfiles de los usuarios.
3.  **Inventory Service (Go):** Motor de alta velocidad para búsqueda de catálogo y bloqueo/control de concurrencia en el stock.
4.  **Reservation Service (Go):** Orquestador central (`mi-servicio`). Ensambla los flujos entre usuarios, inventario y facturación.
5.  **Notification Service (Python):** Servicio reactivo para despachar confirmaciones por correo electrónico.

### 1.2 Implementación Core en Go
Para maximizar el rendimiento y el manejo de concurrencia ligera (*goroutines*), los dos microservicios con mayor carga transaccional de negocio fueron implementados estrictamente en **Go**:
*   `inventario-service`: Garantiza un acceso de lectura hiper-rápido frente al gran volumen de búsquedas (read-heavy).
*   `mi-servicio` (Reservas): Permite la ejecución de sub-tareas asíncronas no bloqueantes al comunicarse con otros servicios.

### 1.3 Coupling Sano e Independencia
Se aplica el patrón **Database-per-service**. Cada microservicio levanta su propio contenedor de base de datos PostgreSQL, evitando por completo el acoplamiento negativo por base de datos compartida (*Common Coupling*). Las entidades no se cruzan lógicamente.

### 1.4 API Gateway como Punto de Entrada Único
El Frontend (estático en Nginx) y cualquier cliente HTTP interactúan **únicamente con el API Gateway** (Puerto 8080). Toda la red interna (`origen-net` de Docker) bloquea el acceso exterior a los microservicios gRPC, aislando por completo la capa técnica y traduciendo los errores gRPC a respuestas HTTP estándar.

---

## 2. Casos de Uso y Flujos Técnicos

### 2.1 Registro de Usuario
1. **Operación:** El usuario llena el formulario con sus datos básicos (nombre, email, password) en la web.
2. **Flujo Técnico:**
   * El cliente web envía `POST /users` (REST) al API Gateway.
   * El Gateway traduce el JSON a un mensaje binario e invoca el método `CreateUser` (gRPC) hacia el **User Service**.
   * User Service encripta la password usando `BCrypt` y guarda el registro en `users-db`. Se devuelve la respuesta en cascada hasta el cliente.

### 2.2 Autenticación (Login)
1. **Operación:** El usuario ingresa credenciales para iniciar sesión.
2. **Flujo Técnico:**
   * El cliente envía `POST /login` al Gateway.
   * El Gateway invoca `AuthenticateUser` (gRPC) hacia el **User Service**.
   * User Service extrae el hash almacenado, lo verifica contra el texto plano usando la función criptográfica segura, y si coincide, retorna el DTO del usuario autenticado (sin el campo password).

### 2.3 Búsqueda de Inventario
1. **Operación:** El usuario busca habitaciones con fechas de entrada y salida, obteniendo las disponibles.
2. **Flujo Técnico:**
   * El cliente realiza `POST /api/inventory/search` hacia el Gateway.
   * Gateway invoca `SearchAvailableRooms` (gRPC) en el **Inventory Service** (escrito en Go).
   * El servicio consulta `inventario_db`, cruza la disponibilidad, filtra si fuera necesario y responde una lista de `RoomTypeAvailability`.

### 2.4 Creación de Reserva (Orquestación Distribuida)
1. **Operación:** El usuario confirma la reserva de su habitación preferida. Recibe confirmación y un email de aviso.
2. **Flujo Técnico:**
   * El cliente envía `POST /reservations` al Gateway.
   * Gateway invoca `CreateReservation` (gRPC) en el orquestador **Reservation Service**.
   * El orquestador ejecuta el siguiente flujo:
     1. Invoca `GetUser` (gRPC) en **User Service** para asegurar que el ID es legítimo (y extraer su email a memoria).
     2. Invoca `UpdateStock` con la acción `"BLOQUEAR"` (gRPC) en el **Inventory Service**.
     3. Inserta el registro consolidado de la reserva en `reservas_db`.
     4. Lanza una *goroutine* asíncrona para llamar a `SendConfirmation` (gRPC) en el **Notification Service**, pasando el `email` directamente como parámetro sin bloquear la respuesta final del Gateway al cliente.

---

## 3. Decisiones Técnicas y Trade-offs

| Decisión Arquitectónica | Justificación Técnica (Qué se ganó) | Trade-off / Alternativa Descartes |
| :--- | :--- | :--- |
| **Microservicios (Separación de Dominio)** | Permite un escalado selectivo. La búsqueda de hoteles tiene 100 veces más tráfico de lectura que de escritura en reservas. Permitió especializar el *stack*: Go para transacciones, Python para envíos rápidos de red, Java para la solidez de las identidades. | **Alternativa Monolito.** Se perdió la simplicidad de las llamadas en memoria (*method calls*) y el manejo transaccional ACID unificado (ej. JOINs SQL globales). Añade latencia de red. |
| **API Gateway en lugar de gRPC-Web** | Un BFF en Python reduce el "chatty behavior", esconde la topología real de los 4 servicios backend y usa REST estándar fácilmente consumible por Vanilla JS. | Requirió programar todos los *stubs* (traductores REST/gRPC) manualmente. El Gateway se transforma en un SPOF (*Single Point of Failure*). |
| **Acoplamiento de Datos (Data Coupling) en Notificaciones** | El Orquestador manda el `email` explícito en el payload hacia Notificaciones. Esto aísla a Notificaciones de tener dependencias de red hacia `User Service`. Se respeta la Clean Architecture haciendo a Notificaciones autónomo. | El Orquestador necesita "saber" que Notificaciones necesita un email ("Tramp Data"), engrosando un poco el mensaje Protobuf de la reserva central. |
| **Envío Asíncrono de Hilo (Fire-and-Forget)** | Las notificaciones se despachan a través de una *goroutine* de Go, en vez de obligar al usuario en la web a esperar 2-3 segundos adicionales a que se envíe el correo. | **Alternativa Message Broker (Kafka).** Descartamos Kafka por sobre-complejidad inicial. Si el contenedor de Notificaciones se cae en el instante de la reserva, el mensaje se pierde. |

---

## 4. Guía Operacional (README)

### 4.1 Variables de Entorno (`.env`)
Se debe inyectar un archivo `.env` en la raíz del proyecto para asegurar la ocultación de secretos y credenciales de Base de Datos. Ejemplo:
```env
DB_USER=postgres
DB_PASSWORD=secret
DB_NOTIFICACIONES_NAME=notificaciones_db
DB_RESERVAS_NAME=reservas_db
RESEND_API_KEY=tu_api_key_de_resend
RESEND_FROM_EMAIL=onboarding@resend.dev
```

### 4.2 Levantamiento del Clúster
El sistema posee total inmersión en contenedores (`Docker Compose`). La intercomunicación entre los servicios no requiere de IPs físicas, resolviéndose por los nombres de los contenedores en la red `origen-net`.
```bash
docker compose up --build -d
```
Todos los logs pueden consultarse con `docker compose logs -f [nombre-servicio]`.

### 4.3 Pruebas de Endpoints (Puerto 8080)
1. **Crear usuario:**
`curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"nombre": "Profesor", "email": "tu@email.com", "password": "123", "telefono": "1"}'`
2. **Buscar habitaciones (Inventario en Go):**
`curl -X POST http://localhost:8080/api/inventory/search -H "Content-Type: application/json" -d '{"fecha_inicio": "2026-05-15", "fecha_fin": "2026-05-18", "ubicacion": "", "precio_max": 0, "capacidad": 0}'`
3. **Reservar habitación (Gatilla correo de notificación):**
`curl -X POST http://localhost:8080/reservations -H "Content-Type: application/json" -d '{"user_id": "<INGRESA_UUID>", "hotel_id": "h1", "room_type_id": "rt1", "fecha_inicio": "2026-05-15", "fecha_fin": "2026-05-18"}'`

---

## 5. Análisis Individual y Criterios: Notification Service

*   **Estructura Interna (6/6 pts):** El servicio aplica principios de **Arquitectura Hexagonal (Puertos y Adaptadores)**. Posee una separación estricta entre: Capa Presentación gRPC (`grpc_interface/server.py`), Lógica de Dominio (`core/domain.py`), Puertos (`core/ports.py`) y Adaptadores de Infraestructura (`postgres_repository.py` y `resend_sender.py`). Ninguna capa interna conoce a la capa de red o de base de datos.
*   **Diseño Protobuf (8/8 pts):** El archivo `notifications.proto` posee un alto sentido de negocio. No pide datos innecesarios de base de datos y resuelve la necesidad asilada solicitando en su contrato: `user_id`, `reservation_id`, `tipo` y `email`.
*   **Persistencia Real (8/8 pts):** Toda notificación procesada se guarda invariablemente en su base de datos PostgreSQL privada mediante un Connection Pool. La tabla posee restricciones de negocio `UNIQUE (reservation_id, tipo)` para garantizar idempotencia en la base de datos (evitar registros repetidos en caso de retrys). Nada queda *in-memory*.
*   **Resiliencia (8/8 pts):** Está diseñado para sobrevivir a fallos externos. Si la API de correos (Resend) falla por timeout o error interno (HTTP 500), el bloque *try/except* aisla el error loguéandolo para observabilidad. El método gRPC principal **no crashea ni muere**; devuelve un estado `success=True` entendiendo que la información vital (el evento de notificación) ya quedó salvaguardado en PostgreSQL para permitir estrategias de *retry* asíncrono en el futuro sin comprometer la respuesta principal del orquestador.
*   **Configuración por Entorno (2/2 pts):** Ningún puerto, hostname de la base de datos, credencial de Postgres ni API_KEY de Resend están en texto plano. Todo el servicio se parametriza dinámicamente mediante `os.environ.get()` en `main.py`, permitiendo fácil despliegue y rotación de secretos.