from prometheus_client import Counter, Histogram

notifications_saved_total = Counter(
    "notifications_saved_total",
    "Notificaciones guardadas correctamente.",
    ["tipo"],
)

notifications_external_total = Counter(
    "notifications_external_total",
    "Resultado del envio por canal externo.",
    ["status"],
)

notifications_errors_total = Counter(
    "notifications_errors_total",
    "Errores del servicio de notificaciones.",
    ["operation", "error_type"],
)

notification_send_duration_seconds = Histogram(
    "notification_send_duration_seconds",
    "Duracion del procesamiento de SendConfirmation.",
)
