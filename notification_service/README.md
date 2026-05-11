# Notification Service

Microservicio encargado de gestionar las notificaciones asíncronas para el Sistema de Gestión de Reservas de Hoteles (Origen X).

Está diseñado siguiendo los principios de la **Arquitectura Hexagonal (Clean Architecture)** y **Test-Driven Development (TDD)**. Expone una interfaz **gRPC** y persiste los eventos en **PostgreSQL**, garantizando alta resiliencia e idempotencia frente a fallos de red o particiones.

## Tecnologías Principales
- **Lenguaje:** Python 3.11+
- **Framework RPC:** gRPC (`grpcio`)
- **Base de Datos:** PostgreSQL (`psycopg` pool)
- **Calidad y Testing:** `pytest`, `mypy`, `ruff`, `black`

## Estructura del Proyecto
```text
notification_service/
├── src/
│   ├── core/            # Entidades de dominio y puertos (interfaces)
│   ├── infrastructure/  # Adaptadores de base de datos (PostgreSQL)
│   ├── grpc_interface/  # Adaptadores de red (Servidor gRPC)
│   ├── proto/           # Definiciones Protobuf compiladas
│   └── main.py          # Punto de entrada de la aplicación
├── tests/               # Pruebas unitarias e integración (TDD)
├── proto/               # Archivo original .proto
├── Dockerfile           # Receta de contenedor
└── requirements.txt     # Dependencias de Python
```

## Prerrequisitos

- **Python 3.11** o superior (para desarrollo local).
- **PostgreSQL** corriendo localmente o en un contenedor.
- **Docker** y **Docker Compose** (opcional, para despliegue).

## Instalación Local y Desarrollo

1. **Clonar y preparar entorno virtual:**
   ```bash
   python3 -m venv venv
   source venv/bin/activate
   ```

2. **Instalar dependencias:**
   ```bash
   pip install -r requirements.txt
   ```

3. **Configurar variables de entorno:**
   Copia el archivo de ejemplo y ajusta las credenciales si es necesario.
   ```bash
   cp .env.example .env
   ```

4. **Compilar los archivos Protobuf (solo si modificas el `.proto`):**
   ```bash
   python -m grpc_tools.protoc -Iproto --python_out=src/proto --grpc_python_out=src/proto --mypy_out=src/proto proto/notifications.proto
   sed -i 's/import notifications_pb2 as notifications__pb2/from proto import notifications_pb2 as notifications__pb2/g' src/proto/notifications_pb2_grpc.py
   ```

## Ejecutar Pruebas (TDD)

El proyecto cuenta con 100% de cobertura lógica en sus pruebas (incluyendo *mocks* para simular fallos de base de datos y validaciones de *timeouts* de red).

```bash
# Ejecutar pruebas con reporte de cobertura
pytest -v --cov=src tests/

# Validar estricto tipado (Static Type Checking)
mypy src tests

# Validar formateo y buenas prácticas
ruff check .
```

## Ejecutar el Servicio

**Opción A: Localmente (Python directo)**
```bash
source venv/bin/activate
python src/main.py
```

**Opción B: Con Docker**
```bash
docker build -t notification-service .
docker run -p 50051:50051 --env-file .env notification-service
```

## API gRPC

El servicio expone la siguiente definición:

**`SendConfirmation`**
Recibe los datos de una reserva y registra la notificación asíncrona ("fire-and-forget").

*Request:*
```json
{
  "user_id": "string",
  "reservation_id": "string",
  "tipo": "string" // Ej: "CONFIRMACION", "CANCELACION"
}
```

*Response:*
```json
{
  "success": true
}
```

### Características de Resiliencia
- **Idempotencia:** Si ocurre un reintento por problemas de red (`reservation_id` + `tipo` duplicados), el servicio responde un éxito sin insertar duplicados.
- **Timeout Awareness:** Si el cliente cancela la solicitud anticipadamente (por *deadline/timeout*), el servicio corta la ejecución instantáneamente para ahorrar CPU y conexiones a la BD.