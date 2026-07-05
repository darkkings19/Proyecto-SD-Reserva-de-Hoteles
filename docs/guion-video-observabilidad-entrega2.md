## 0:00 - 0:25 Introduccion

Hola, mi bloque corresponde a observabilidad dentro del sistema Origen X, una aplicacion distribuida para reserva de hoteles.

El sistema tiene frontend, API Gateway y microservicios de usuarios, reservas, inventario y notificaciones. Mi modulo principal fue usuarios, pero para observabilidad tambien extendi la implementacion al flujo distribuido completo, porque una reserva real pasa por varios servicios.

## 0:25 - 0:55 Que es observabilidad y por que use 3 pilares

Para esta entrega implemente observabilidad usando los tres pilares: metricas, logs y trazas.

Use Prometheus para metricas, porque permite saber cuantos eventos ocurren y si los servicios estan vivos.

Use Loki para logs, porque permite centralizar los mensajes que generan los contenedores.

Use Tempo con OpenTelemetry para trazas, porque permite ver el recorrido de una peticion entre microservicios.

Grafana lo use como punto central para visualizar los tres pilares en dashboards y en Explore.

## 0:55 - 1:25 Que tuve que implementar

En usuarios agregue metricas de usuarios creados, logins exitosos, logins fallidos, sesiones activas, validacion de tokens y errores.

En reservas, inventario y notificaciones agregue metricas propias de cada servicio.

Tambien agregue Loki y Promtail para logs, Tempo y OpenTelemetry Collector para trazas, y configure Grafana con tres dashboards: uno de usuarios, uno del sistema completo y uno de los tres pilares.

Ademas, el API Gateway y los servicios propagan contexto de trazas para que una peticion de reserva pueda verse conectada de punta a punta.


## 2:35 - 2:55 Fallos y limitaciones

Si un servicio cae, Prometheus deja de mostrarlo como `UP`. Si hay errores de negocio, como login fallido o reserva invalida, quedan reflejados en metricas y logs.

Una limitacion es que esto esta preparado para entorno local con Docker Compose. En produccion faltarian alertas, seguridad mas fuerte, retencion configurada y una estrategia mas robusta de logs.

## 2:55 - 3:00 Cierre

En resumen, el sistema ahora no solo ejecuta reservas, sino que permite observar que ocurre internamente con metricas, logs y trazas distribuidas.

