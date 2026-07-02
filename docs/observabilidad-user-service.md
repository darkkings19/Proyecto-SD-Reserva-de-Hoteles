# Bloque: Observabilidad del modulo de usuarios

## Descripcion del bloque

Este bloque agrega observabilidad al `user-service`, el microservicio responsable de crear usuarios, autenticar credenciales, consultar perfiles y actualizar datos basicos. El servicio participa en dos flujos del sistema distribuido:

- `frontend` -> `api_gateway` -> `user-service` -> `users-db`, para registro y login.
- `reservation_service` -> `user-service`, para validar que el usuario exista antes de crear una reserva.

La observabilidad implementada expone health checks, metricas Prometheus, logs de eventos relevantes y un dashboard Grafana. Esto permite revisar si el servicio esta disponible, si la base de datos responde y como se comportan las operaciones principales del modulo.

## Decision tecnica

Se eligio usar Spring Boot Actuator con Micrometer y Prometheus para exponer metricas en `/actuator/prometheus`, y Grafana para visualizarlas en un panel fijo.

Alternativa descartada: usar solo logs en consola. Los logs son utiles para depurar eventos puntuales, pero no permiten consultar facilmente tendencias como cantidad de logins fallidos, usuarios creados o errores por tipo. Prometheus permite recolectar esas metricas periodicamente y consultarlas con nombres estables, y Grafana permite mostrarlas en dashboard con rango de fecha y hora seleccionable.

Metricas especificas del dominio:

- `users_created_total`: usuarios creados correctamente.
- `users_get_by_id_total`: consultas exitosas por ID.
- `users_updated_total`: perfiles actualizados.
- `users_login_success_total`: autenticaciones exitosas.
- `users_login_failed_total`: autenticaciones fallidas.
- `users_domain_errors_total`: errores de dominio etiquetados por operacion y tipo de error.
- `users_logout_success_total`: cierres de sesion exitosos.
- `users_token_validated_total`: validaciones exitosas de token.
- `users_active_sessions`: sesiones activas en este momento.

Las metricas con sufijo `_total` son contadores acumulativos desde que arranco el servicio. Para ver un periodo especifico, Grafana usa consultas PromQL con el rango seleccionado por el usuario, por ejemplo:

- `increase(users_created_total[$__range])`: usuarios creados en el periodo seleccionado.
- `increase(users_login_success_total[$__range])`: logins exitosos en el periodo seleccionado.
- `increase(users_login_failed_total[$__range])`: logins fallidos en el periodo seleccionado.

## Comportamiento ante fallos

Si se intenta registrar un email repetido, el servicio responde con error gRPC `ALREADY_EXISTS`, registra un log `WARN` e incrementa `users_domain_errors_total{operation="create_user", error_type="email_already_exists"}`.

Si el login falla por email inexistente o password invalida, el servicio responde con `UNAUTHENTICATED`, registra un log `WARN` e incrementa `users_login_failed_total` y `users_domain_errors_total{operation="authenticate", error_type="invalid_credentials"}`.

Si reservas consulta un usuario inexistente, `user-service` responde con `NOT_FOUND`, registra un log `WARN` e incrementa `users_domain_errors_total{operation="get_user_by_id", error_type="user_not_found"}`. El flujo de reserva se degrada de forma controlada porque no crea la reserva con un usuario invalido.

Si la base de datos de usuarios deja de responder, el endpoint `/actuator/health` deja de reportar estado saludable para el componente de base de datos. Prometheus sigue intentando consultar metricas y el operador puede detectar el problema por health check y logs.

Si el token JWT expira, esta mal firmado, no tiene sesion asociada o la sesion fue revocada por logout, `user-service` responde `UNAUTHENTICATED`, registra un log `WARN` e incrementa `users_domain_errors_total` con `operation="token"`. La sesion deja de contarse en `users_active_sessions` cuando expira o cuando se ejecuta logout.

## Trade-offs y limitaciones

La observabilidad esta enfocada en el modulo de usuarios, no en todos los microservicios. Esto reduce el alcance y permite mantener una implementacion consistente con la responsabilidad individual del bloque, pero no reemplaza una estrategia completa de observabilidad distribuida.

No se implemento tracing distribuido con OpenTelemetry. Por eso se puede observar lo que ocurre dentro de `user-service`, pero no seguir automaticamente una misma request desde frontend hasta reservas, inventario y notificaciones.

Prometheus se configura con un scrape simple cada 5 segundos. Es suficiente para la demostracion y desarrollo, pero en produccion se ajustarian retencion, alertas, autenticacion y dashboards.

Para poder medir usuarios activos se agrego estado de sesion en PostgreSQL. Esto permite logout y conteo real de sesiones vigentes, pero introduce una decision de diseno: el login ya no es completamente stateless. Se acepto este trade-off porque un JWT puramente stateless no permite saber con certeza cuantos usuarios siguen activos ni revocar sesiones antes de la expiracion.

## Como demostrarlo

Levantar el sistema:

```powershell
docker compose up -d --build user-service api_gateway prometheus grafana
```

Ver health del servicio:

```powershell
Invoke-RestMethod http://localhost:8081/actuator/health
```

Crear usuario:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/users" -Method Post -Headers @{"Content-Type"="application/json"} -Body '{"nombre":"Usuario Observabilidad","email":"obs@test.com","password":"pass123","telefono":"123456"}'
```

Login exitoso:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/login" -Method Post -Headers @{"Content-Type"="application/json"} -Body '{"email":"obs@test.com","password":"pass123"}'
```

Login fallido:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/login" -Method Post -Headers @{"Content-Type"="application/json"} -Body '{"email":"obs@test.com","password":"incorrecta"}'
```

Validar token:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/validate-token" -Method Post -Headers @{"Content-Type"="application/json"} -Body '{"access_token":"PEGAR_TOKEN_DEL_LOGIN"}'
```

Logout:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/logout" -Method Post -Headers @{"Content-Type"="application/json"} -Body '{"access_token":"PEGAR_TOKEN_DEL_LOGIN"}'
```

Consultar metricas directamente:

```powershell
Invoke-RestMethod http://localhost:8081/actuator/prometheus
```

Abrir Prometheus:

```text
http://localhost:9091
```

Abrir Grafana:

```text
http://localhost:3001
```

Credenciales:

```text
usuario: admin
password: admin
```

Dashboard:

```text
Origen X / Observabilidad User Service
```

En Grafana, el selector de tiempo superior derecho permite ver metricas por rango como "Last 5 minutes", "Today" o un rango personalizado con dia y hora.

Consultas utiles en Prometheus:

```text
users_created_total
users_login_success_total
users_login_failed_total
users_domain_errors_total
users_active_sessions
up{job="user-service"}
```

## Matriz de cumplimiento de la rubrica

| Criterio | Evidencia en esta implementacion |
| --- | --- |
| Descripcion del bloque | El bloque esta acotado al `user-service`, pero se explica su rol en los flujos `frontend` -> `api_gateway` -> `user-service` -> `users-db` y `reservation_service` -> `user-service`. |
| Decisiones tecnicas | Se documenta la decision de usar Actuator + Micrometer + Prometheus + Grafana y se descarta usar solo logs, porque no permite consultar metricas acumuladas, tendencias ni paneles por rango de tiempo. |
| Comportamiento ante fallos | Se describen email duplicado, credenciales invalidas, usuario inexistente, base de datos no disponible y token JWT invalido/expirado/revocado. Cada caso tiene error gRPC, log y/o metrica observable. |
| Trade-offs y limitaciones | Se reconoce que el alcance es observabilidad del modulo de usuarios, no tracing distribuido completo ni monitoreo productivo con alertas y retencion avanzada. Tambien se reconoce que medir usuarios activos requiere persistir sesiones. |
| Bloque funciona correctamente | El endpoint `/actuator/prometheus` expone metricas del dominio; Prometheus las recolecta desde `user-service:8080`; Grafana muestra totales, periodos seleccionables y sesiones activas. |
| Calidad del codigo | La instrumentacion queda separada en `UserMetrics` y `UserSessionMetrics`; la logica de negocio sigue en `UserServiceImpl`; la configuracion queda en `application.yml`, `docker-compose.yml`, `observability/prometheus.yml` y provisioning de Grafana. |
| Coherencia documento-codigo | Las metricas documentadas corresponden a las expuestas por Prometheus y usadas por Grafana: `users_created_total`, `users_login_success_total`, `users_login_failed_total`, `users_domain_errors_total`, `users_logout_success_total`, `users_token_validated_total` y `users_active_sessions`. |

## Pruebas automatizadas

Se agregaron pruebas unitarias para asegurar que las metricas suben en casos felices y casos de borde:

- Creacion exitosa incrementa `users.created`.
- Email duplicado incrementa `users.domain.errors` con `operation=create_user`.
- Login exitoso incrementa `users.login.success`.
- Login fallido incrementa `users.login.failed` y `users.domain.errors` con `operation=authenticate`.

Estas pruebas validan los nombres internos de Micrometer. En Prometheus esos nombres se exponen con formato Prometheus y sufijo `_total`, por ejemplo `users.login.success` se consulta como `users_login_success_total`.
