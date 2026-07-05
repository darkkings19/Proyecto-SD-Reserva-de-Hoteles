# Observabilidad 3 Pilares - Origen X

Esta rama completa la observabilidad del sistema con los tres pilares clasicos:

1. Metricas: Prometheus + Grafana.
2. Logs: Loki + Promtail + Grafana.
3. Trazas distribuidas: OpenTelemetry + Collector + Tempo + Grafana.

## Componentes

- Prometheus: recolecta metricas de `user-service`, `reservation_service`, `inventario_service` y `notification_service`.
- Loki: almacena logs centralizados de los contenedores Docker.
- Promtail: descubre contenedores y envia sus logs a Loki con etiquetas como `service` y `container`.
- Tempo: almacena trazas distribuidas.
- OpenTelemetry Collector: recibe trazas OTLP desde los servicios y las envia a Tempo.
- Grafana: visualiza metricas, logs y trazas.

## URLs

- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- Prometheus: http://localhost:9091
- Grafana: http://localhost:3001
- Loki: http://localhost:3100
- Tempo: http://localhost:3200

Credenciales de Grafana:

- Usuario: `admin`
- Password: `admin`

## Como demostrarlo

1. Levantar el sistema:

```bash
docker compose up -d --build
```

2. Generar actividad:

- Crear usuario.
- Iniciar sesion.
- Buscar habitaciones.
- Crear una reserva.

3. En Grafana abrir el dashboard:

`Origen X - Observabilidad 3 Pilares`

4. Mostrar los pilares:

- Metricas: panel de Prometheus y dashboards existentes.
- Logs: panel de Loki con logs por servicio.
- Trazas: panel de Tempo o Explore > Tempo.

## Explicacion corta

Prometheus responde si algo esta pasando: errores, latencia, cantidad de operaciones.
Loki permite ver los mensajes exactos de los servicios.
Tempo permite seguir una peticion entre servicios, por ejemplo:

`api-gateway -> reservation-service -> user-service -> inventory-service -> notification-service`

Asi, si una reserva falla, se puede mirar primero la metrica, luego el log y finalmente la traza para ubicar donde fallo.
