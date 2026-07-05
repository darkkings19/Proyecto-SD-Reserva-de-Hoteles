# Evaluacion segun rubrica - Entrega 2 - Observabilidad

Este documento evalua el estado actual del proyecto respecto de la rubrica individual del bloque de observabilidad. La evaluacion se basa en el codigo y configuracion presentes en `main`, mas las verificaciones ejecutadas localmente.

## Resumen de puntaje estimado

| Categoria | Maximo | Estimado actual | Estado |
| --- | ---: | ---: | --- |
| Documento de tu bloque | 40 pts | 40 pts | Cubierto con `documento-bloque-observabilidad-entrega2.md` |
| Implementacion de tu bloque | 30 pts | 29 pts | Funcional y coherente; quedan limitaciones productivas normales |
| Video individual | nota 1-7 | No evaluable aqui | Se entrega guion sugerido |

Nota estimada sin video: muy alta. Si el video demuestra el flujo con claridad, el bloque queda bien defendido.

## Documento de tu bloque - 40 pts

| Criterio | Maximo | Estimado | Evidencia |
| --- | ---: | ---: | --- |
| Descripcion del bloque | 8 | 8 | El documento explica que observabilidad permite mirar metricas, logs y trazas del sistema de reservas. Describe como encaja con frontend, API Gateway, usuarios, reservas, inventario y notificaciones. |
| Decisiones tecnicas | 12 | 12 | Se documentan decisiones especificas: usar los 3 pilares, centralizar visualizacion en Grafana, usar OpenTelemetry Collector antes de Tempo, usar Promtail para leer logs Docker y mantener sesiones para usuarios activos. Cada decision incluye alternativa descartada. |
| Comportamiento ante fallos | 12 | 12 | Se explican fallos de Prometheus, Loki/Promtail, Tempo/Collector, servicios caidos, base de datos y errores de negocio como login fallido o reserva invalida. |
| Trade-offs y limitaciones | 8 | 8 | Se reconocen limitaciones reales: entorno local, sin alertas productivas, retencion simple, logs no completamente estructurados, secretos de desarrollo y observabilidad no bloqueante. |

Resultado documento: 40/40 si se entrega el documento nuevo y se defiende de forma consistente.

## Implementacion de tu bloque - 30 pts

| Criterio | Maximo | Estimado | Evidencia |
| --- | ---: | ---: | --- |
| El bloque funciona correctamente | 12 | 12 | Prometheus ve los cuatro servicios como `UP`; Loki recibe labels y servicios; Tempo recibe trazas; una reserva puede verse como traza distribuida entre gateway, reservas, usuarios, inventario y notificaciones. |
| Calidad del codigo | 10 | 9 | La instrumentacion esta separada por archivos y respeta la estructura: `api-gateway/observability.py`, `notification_service/src/tracing.py`, metricas en servicios Go, metricas de usuario en clases dedicadas, provisioning en `observability/`. Se descuenta levemente porque es una implementacion academica/local con configuracion simple y algunos valores de desarrollo. |
| Coherencia con el documento | 8 | 8 | El documento describe exactamente Prometheus, Loki, Tempo, OpenTelemetry Collector, Promtail, Grafana, dashboards y metricas existentes. |

Resultado implementacion: 29/30.

## Verificaciones realizadas

| Verificacion | Resultado |
| --- | --- |
| `docker compose config --quiet` | Correcto |
| Prometheus targets | `user-service`, `reservation-service`, `inventory-service`, `notification-service` en `UP` |
| Loki labels | Existen labels `compose_project`, `container`, `service`, `stream` |
| Loki services | Aparecen servicios como `api_gateway`, `reservation_service`, `user-service`, `notification_service`, `inventario_service` |
| Tempo search | Devuelve trazas recientes |
| `mvn test` en `user-service` | Correcto, 4 tests pasan |
| `go test ./...` en `inventario-service` | Correcto |
| `go test ./...` en `mi-servicio` | Correcto |
| `python -m py_compile` en gateway y notificaciones | Correcto |

## Riesgos o puntos que debe explicar el video

- Loki no tiene una interfaz visual comoda propia; por eso se muestra mejor desde Grafana Explore.
- Tempo por si solo entrega datos tecnicos por API; la vista clara de trazas esta en Grafana.
- Prometheus si tiene interfaz propia, pero Grafana permite cruzar metricas, logs y trazas en un mismo lugar.
- No se implementaron alertas porque la rubrica pide observabilidad del bloque, no alertamiento productivo.
- El entorno es local con Docker Compose; en produccion se ajustarian retencion, seguridad, secretos, volumenes y alertas.

## Recomendacion para maximizar puntaje

En el video conviene defender la decision asi:

1. Explicar que el bloque elegido fue observabilidad y que el modulo principal asignado fue usuarios.
2. Decir que inicialmente se observo usuarios, pero luego se extendio a los servicios principales para demostrar un flujo distribuido real.
3. Mostrar una accion concreta: login o creacion de reserva.
4. Mostrar los tres pilares en Grafana: metricas, logs y trazas.
5. Cerrar explicando una decision tecnica y una limitacion real.

