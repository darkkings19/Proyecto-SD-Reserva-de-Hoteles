package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "github.com/darkkings19/mi-servicio/pb"
	_ "github.com/lib/pq"
)

const (
	sagaStarted                          = "STARTED"
	sagaUserValidated                    = "USER_VALIDATED"
	sagaStockLocked                      = "STOCK_LOCKED"
	sagaReservationCreated               = "RESERVATION_CREATED"
	sagaNotificationDispatching          = "NOTIFICATION_DISPATCHING"
	sagaCompleted                        = "COMPLETED"
	sagaCompletedWithNotificationFailure = "COMPLETED_WITH_NOTIFICATION_FAILURE"
	sagaCompensating                     = "COMPENSATING"
	sagaCompensated                      = "COMPENSATED"
	sagaCompensationFailed               = "COMPENSATION_FAILED"
	sagaFailed                           = "FAILED"

	sagaUserValidationFailed    = "USER_VALIDATION_FAILED"
	sagaStockLockFailed         = "STOCK_LOCK_FAILED"
	sagaStockUnavailable        = "STOCK_UNAVAILABLE"
	sagaReservationInsertFailed = "RESERVATION_INSERT_FAILED"
)

var (
	reservationsCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reservations_created_total",
		Help: "Reservas creadas correctamente.",
	})
	reservationsListTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reservations_list_total",
		Help: "Listados de reservas consultados correctamente.",
	})
	reservationsGetTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reservations_get_total",
		Help: "Reservas consultadas por ID correctamente.",
	})
	reservationsFailuresTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "reservations_failures_total",
		Help: "Fallos del flujo de reservas etiquetados por etapa.",
	}, []string{"stage"})
	reservationsNotificationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "reservations_notifications_total",
		Help: "Resultado de notificaciones disparadas desde reservas.",
	}, []string{"status"})
	reservationCreationDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "reservations_create_duration_seconds",
		Help:    "Duracion de la creacion de reservas.",
		Buckets: prometheus.DefBuckets,
	})
	reservationSagaTransitionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "reservation_saga_transitions_total",
		Help: "Transiciones registradas por la saga de reservas.",
	}, []string{"state", "result"})
	reservationSagaCompensationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "reservation_saga_compensations_total",
		Help: "Compensaciones de inventario ejecutadas por la saga de reservas.",
	}, []string{"result"})
	reservationsIdempotentReplaysTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "reservations_idempotent_replays_total",
		Help: "Veces que se devolvio una reserva existente en vez de crear una nueva por idempotency_key repetida.",
	})
)

func init() {
	prometheus.MustRegister(
		reservationsCreatedTotal,
		reservationsListTotal,
		reservationsGetTotal,
		reservationsFailuresTotal,
		reservationsNotificationsTotal,
		reservationCreationDuration,
		reservationSagaTransitionsTotal,
		reservationSagaCompensationsTotal,
		reservationsIdempotentReplaysTotal,
	)
	reservationsFailuresTotal.WithLabelValues("user_validation").Add(0)
	reservationsFailuresTotal.WithLabelValues("inventory_lock").Add(0)
	reservationsFailuresTotal.WithLabelValues("inventory_no_stock").Add(0)
	reservationsFailuresTotal.WithLabelValues("database_insert").Add(0)
	reservationsFailuresTotal.WithLabelValues("list_database").Add(0)
	reservationsFailuresTotal.WithLabelValues("get_database").Add(0)
	reservationsFailuresTotal.WithLabelValues("get_not_found").Add(0)
	reservationsNotificationsTotal.WithLabelValues("success").Add(0)
	reservationsNotificationsTotal.WithLabelValues("failed").Add(0)
	for _, state := range []string{
		sagaStarted,
		sagaUserValidated,
		sagaStockLocked,
		sagaReservationCreated,
		sagaNotificationDispatching,
		sagaCompleted,
		sagaCompletedWithNotificationFailure,
		sagaCompensating,
		sagaCompensated,
		sagaCompensationFailed,
		sagaFailed,
	} {
		reservationSagaTransitionsTotal.WithLabelValues(state, "success").Add(0)
		reservationSagaTransitionsTotal.WithLabelValues(state, "failed").Add(0)
	}
	reservationSagaCompensationsTotal.WithLabelValues("success").Add(0)
	reservationSagaCompensationsTotal.WithLabelValues("failed").Add(0)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func startMetricsServer(port string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("[Reservas] Metricas Prometheus en :%s/metrics", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Printf("[Reservas] Error en servidor de metricas: %v", err)
		}
	}()
}

func initTracer(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return provider.Shutdown, nil
}

func generateWorkflowID(prefix string) string {
	var randomBytes [4]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s-%s-%x", prefix, time.Now().UTC().Format("20060102150405"), randomBytes)
}

func initializeReservationSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS reservations (
			reservation_id VARCHAR(80) PRIMARY KEY,
			user_id VARCHAR(80) NOT NULL,
			hotel_id VARCHAR(80) NOT NULL,
			room_type_id VARCHAR(80) NOT NULL,
			fecha_inicio VARCHAR(20) NOT NULL,
			fecha_fin VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL,
			monto_total NUMERIC(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE reservations ADD COLUMN IF NOT EXISTS saga_id VARCHAR(80);`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_user_created_at ON reservations (user_id, created_at DESC);`,
		`ALTER TABLE reservations ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(120);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_reservations_idempotency_key ON reservations (idempotency_key) WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';`,
		`CREATE TABLE IF NOT EXISTS reservation_sagas (
			saga_id VARCHAR(80) PRIMARY KEY,
			reservation_id VARCHAR(80) NOT NULL,
			user_id VARCHAR(80) NOT NULL,
			hotel_id VARCHAR(80) NOT NULL,
			room_type_id VARCHAR(80) NOT NULL,
			status VARCHAR(80) NOT NULL,
			current_step VARCHAR(80) NOT NULL,
			compensation_status VARCHAR(30) NOT NULL DEFAULT 'NONE',
			error_message TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS reservation_saga_events (
			id SERIAL PRIMARY KEY,
			saga_id VARCHAR(80) NOT NULL REFERENCES reservation_sagas(saga_id) ON DELETE CASCADE,
			step VARCHAR(80) NOT NULL,
			result VARCHAR(30) NOT NULL,
			detail TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reservation_saga_events_saga_created_at ON reservation_saga_events (saga_id, created_at);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

type reservationServer struct {
	pb.UnimplementedReservationServiceServer
	userClient         pb.UserServiceClient
	inventoryClient    pb.InventoryServiceClient
	notificationClient pb.NotificationServiceClient
	db                 *sql.DB

	userBreaker         *CircuitBreaker
	inventoryBreaker    *CircuitBreaker
	notificationBreaker *CircuitBreaker
}

func (s *reservationServer) startSaga(ctx context.Context, sagaID string, reservationID string, req *pb.CreateReservationRequest) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO reservation_sagas (saga_id, reservation_id, user_id, hotel_id, room_type_id, status, current_step)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, sagaID, reservationID, req.UserId, req.HotelId, req.RoomTypeId, sagaStarted)
	if err != nil {
		return err
	}
	s.recordSagaTransition(ctx, sagaID, sagaStarted, "success", "Saga de reserva iniciada")
	return nil
}

func (s *reservationServer) recordSagaTransition(ctx context.Context, sagaID string, state string, result string, detail string) {
	log.Printf("[Saga] saga_id=%s state=%s result=%s detail=%s", sagaID, state, result, detail)
	reservationSagaTransitionsTotal.WithLabelValues(state, result).Inc()

	compensationStatus := "NONE"
	if state == sagaCompensating {
		compensationStatus = "RUNNING"
	} else if state == sagaCompensated {
		compensationStatus = "DONE"
	} else if state == sagaCompensationFailed {
		compensationStatus = "FAILED"
	}

	if _, err := s.db.ExecContext(ctx, `
		UPDATE reservation_sagas
		SET status = $2,
			current_step = $2,
			compensation_status = CASE WHEN $3 <> 'NONE' THEN $3 ELSE compensation_status END,
			error_message = CASE WHEN $4 = 'failed' THEN $5 ELSE error_message END,
			updated_at = CURRENT_TIMESTAMP
		WHERE saga_id = $1
	`, sagaID, state, compensationStatus, result, detail); err != nil {
		log.Printf("[Saga] saga_id=%s error=update_state_failed detail=%v", sagaID, err)
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO reservation_saga_events (saga_id, step, result, detail)
		VALUES ($1, $2, $3, $4)
	`, sagaID, state, result, detail); err != nil {
		log.Printf("[Saga] saga_id=%s error=insert_event_failed detail=%v", sagaID, err)
	}
}

func (s *reservationServer) failSaga(ctx context.Context, sagaID string, state string, detail string) {
	s.recordSagaTransition(ctx, sagaID, state, "failed", detail)
	s.recordSagaTransition(ctx, sagaID, sagaFailed, "failed", detail)
}

func (s *reservationServer) compensateInventory(sagaID string, roomTypeID string) bool {
	baseCtx := context.Background()
	s.recordSagaTransition(baseCtx, sagaID, sagaCompensating, "success", "Liberando stock por fallo posterior al bloqueo")

	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(baseCtx, 3*time.Second)
		res, err := Execute(s.inventoryBreaker, func() (*pb.UpdateStockResponse, error) {
			return s.inventoryClient.UpdateStock(ctx, &pb.UpdateStockRequest{
				RoomTypeId: roomTypeID,
				Cantidad:   1,
				Accion:     "LIBERAR",
			})
		})
		cancel()

		if err == nil && res.GetStatus() {
			reservationSagaCompensationsTotal.WithLabelValues("success").Inc()
			s.recordSagaTransition(baseCtx, sagaID, sagaCompensated, "success", fmt.Sprintf("Stock liberado en intento %d", attempt))
			return true
		}

		detail := fmt.Sprintf("Intento %d fallo liberando stock: %v", attempt, err)
		if err == nil {
			detail = fmt.Sprintf("Intento %d no libero stock: status=false", attempt)
		}
		log.Printf("[Saga] saga_id=%s compensation_attempt=%d result=failed detail=%s", sagaID, attempt, detail)
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}

	reservationSagaCompensationsTotal.WithLabelValues("failed").Inc()
	s.recordSagaTransition(baseCtx, sagaID, sagaCompensationFailed, "failed", "No se pudo liberar stock despues de 3 intentos")
	return false
}

func (s *reservationServer) GetReservation(ctx context.Context, req *pb.GetReservationRequest) (*pb.GetReservationResponse, error) {
	log.Printf("[Reservas] Buscando reserva %s...", req.ReservationId)

	query := "SELECT reservation_id, user_id, hotel_id, room_type_id, status, monto_total, fecha_inicio, fecha_fin FROM reservations WHERE reservation_id = $1"
	var res pb.Reservation
	err := s.db.QueryRowContext(ctx, query, req.ReservationId).Scan(
		&res.ReservationId, &res.UserId, &res.HotelId, &res.RoomTypeId,
		&res.Status, &res.MontoTotal, &res.FechaInicio, &res.FechaFin)

	if err == sql.ErrNoRows {
		reservationsFailuresTotal.WithLabelValues("get_not_found").Inc()
		return nil, status.Errorf(codes.NotFound, "reserva no encontrada")
	} else if err != nil {
		reservationsFailuresTotal.WithLabelValues("get_database").Inc()
		return nil, status.Errorf(codes.Internal, "error en base de datos: %v", err)
	}

	reservationsGetTotal.Inc()
	return &pb.GetReservationResponse{Reservation: &res}, nil
}

func (s *reservationServer) ListReservations(ctx context.Context, req *pb.ListReservationsRequest) (*pb.ListReservationsResponse, error) {
	log.Println("[Reservas] Obteniendo lista de reservas...")

	if req.UserId == "" {
		log.Println("[Reservas] ListReservations sin user_id: retornando lista vacia")
		reservationsListTotal.Inc()
		return &pb.ListReservationsResponse{Reservations: []*pb.Reservation{}}, nil
	}

	query := "SELECT reservation_id, user_id, hotel_id, room_type_id, status, monto_total, fecha_inicio, fecha_fin FROM reservations WHERE user_id = $1 ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query, req.UserId)
	if err != nil {
		log.Printf("[Reservas] Error leyendo base de datos: %v", err)
		reservationsFailuresTotal.WithLabelValues("list_database").Inc()
		return nil, status.Errorf(codes.Internal, "error al consultar la base de datos")
	}
	defer rows.Close()

	var list []*pb.Reservation
	for rows.Next() {
		var res pb.Reservation
		if err := rows.Scan(&res.ReservationId, &res.UserId, &res.HotelId, &res.RoomTypeId, &res.Status, &res.MontoTotal, &res.FechaInicio, &res.FechaFin); err != nil {
			log.Printf("[Reservas] Error escaneando fila: %v", err)
			continue
		}
		list = append(list, &res)
	}

	reservationsListTotal.Inc()
	return &pb.ListReservationsResponse{Reservations: list}, nil
}

func (s *reservationServer) CreateReservation(ctx context.Context, req *pb.CreateReservationRequest) (*pb.CreateReservationResponse, error) {
	timer := prometheus.NewTimer(reservationCreationDuration)
	defer timer.ObserveDuration()

	if req.IdempotencyKey != "" {
		var existingID, existingStatus string
		var existingMonto float64
		err := s.db.QueryRowContext(ctx,
			`SELECT reservation_id, status, monto_total FROM reservations WHERE idempotency_key = $1`,
			req.IdempotencyKey,
		).Scan(&existingID, &existingStatus, &existingMonto)
		if err == nil {
			reservationsIdempotentReplaysTotal.Inc()
			log.Printf("[Reservas] idempotency_key=%s ya fue usada; devolviendo reserva %s sin repetir la saga", req.IdempotencyKey, existingID)
			return &pb.CreateReservationResponse{
				ReservationId: existingID,
				Status:        existingStatus,
				MontoTotal:    existingMonto,
			}, nil
		} else if err != sql.ErrNoRows {
			log.Printf("[Reservas] error verificando idempotency_key=%s: %v (se continua creando la reserva)", req.IdempotencyKey, err)
		}
	}

	sagaID := generateWorkflowID("saga")
	reservationId := generateWorkflowID("res")
	montoTotal := 150.50

	log.Printf("[Reservas] saga_id=%s reservation_id=%s iniciando creacion de reserva para usuario: %s", sagaID, reservationId, req.UserId)
	if err := s.startSaga(ctx, sagaID, reservationId, req); err != nil {
		log.Printf("[Saga] saga_id=%s result=failed detail=no se pudo persistir el inicio de la saga: %v", sagaID, err)
		reservationsFailuresTotal.WithLabelValues("database_insert").Inc()
		return nil, status.Errorf(codes.Internal, "no se pudo iniciar la saga de reserva")
	}

	log.Printf("[Saga] saga_id=%s step=user_validation action=calling_user_service user_id=%s", sagaID, req.UserId)
	userCtx, userCancel := context.WithTimeout(ctx, 5*time.Second)
	userRes, errUser := Execute(s.userBreaker, func() (*pb.UserResponse, error) {
		return s.userClient.GetUser(userCtx, &pb.GetUserRequest{Id: req.UserId})
	})
	userCancel()
	if errUser != nil {
		log.Printf("[Reservas] saga_id=%s error al validar usuario: %v", sagaID, errUser)
		reservationsFailuresTotal.WithLabelValues("user_validation").Inc()
		s.failSaga(context.WithoutCancel(ctx), sagaID, sagaUserValidationFailed, status.Convert(errUser).Message())
		var circuitErr *ErrCircuitOpen
		if errors.As(errUser, &circuitErr) {
			return nil, status.Errorf(codes.Unavailable, "Servicio de usuarios no disponible en este momento")
		}
		return nil, status.Errorf(codes.Unauthenticated, "No se pudo validar el usuario: %v", status.Convert(errUser).Message())
	}
	userEmail := userRes.GetUser().GetEmail()
	s.recordSagaTransition(ctx, sagaID, sagaUserValidated, "success", fmt.Sprintf("Usuario validado: %s", userRes.User.Nombre))

	log.Printf("[Saga] saga_id=%s step=inventory_lock action=calling_inventory_service room_type_id=%s", sagaID, req.RoomTypeId)
	invCtx, invCancel := context.WithTimeout(ctx, 5*time.Second)
	invRes, errInv := Execute(s.inventoryBreaker, func() (*pb.UpdateStockResponse, error) {
		return s.inventoryClient.UpdateStock(invCtx, &pb.UpdateStockRequest{
			RoomTypeId: req.RoomTypeId,
			Cantidad:   1,
			Accion:     "BLOQUEAR",
		})
	})
	invCancel()
	if errInv != nil {
		log.Printf("[Reservas] saga_id=%s error al bloquear inventario: %v", sagaID, errInv)
		reservationsFailuresTotal.WithLabelValues("inventory_lock").Inc()
		s.failSaga(context.WithoutCancel(ctx), sagaID, sagaStockLockFailed, status.Convert(errInv).Message())
		var circuitErr *ErrCircuitOpen
		if errors.As(errInv, &circuitErr) {
			return nil, status.Errorf(codes.Unavailable, "Servicio de inventario no disponible en este momento")
		}
		return nil, status.Errorf(codes.ResourceExhausted, "No se pudo asegurar la habitacion: %v", status.Convert(errInv).Message())
	}
	if !invRes.Status {
		reservationsFailuresTotal.WithLabelValues("inventory_no_stock").Inc()
		s.failSaga(context.WithoutCancel(ctx), sagaID, sagaStockUnavailable, "Inventario respondio status=false")
		return nil, status.Errorf(codes.ResourceExhausted, "No hay stock disponible")
	}
	s.recordSagaTransition(ctx, sagaID, sagaStockLocked, "success", fmt.Sprintf("Stock bloqueado para room_type_id=%s", req.RoomTypeId))

	var idempotencyKeyParam interface{}
	if req.IdempotencyKey != "" {
		idempotencyKeyParam = req.IdempotencyKey
	}

	query := `INSERT INTO reservations (reservation_id, user_id, hotel_id, room_type_id, fecha_inicio, fecha_fin, status, monto_total, saga_id, idempotency_key)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, errDb := s.db.ExecContext(ctx, query,
		reservationId, req.UserId, req.HotelId, req.RoomTypeId,
		req.FechaInicio, req.FechaFin, "CONFIRMADA", montoTotal, sagaID, idempotencyKeyParam)
	if errDb != nil {
		log.Printf("[Reservas] saga_id=%s error al guardar en PostgreSQL: %v", sagaID, errDb)
		reservationsFailuresTotal.WithLabelValues("database_insert").Inc()
		s.recordSagaTransition(context.WithoutCancel(ctx), sagaID, sagaReservationInsertFailed, "failed", errDb.Error())
		s.compensateInventory(sagaID, req.RoomTypeId)
		s.recordSagaTransition(context.Background(), sagaID, sagaFailed, "failed", "Reserva no persistida; se ejecuto compensacion de inventario")
		return nil, status.Errorf(codes.Internal, "error interno al guardar la reserva")
	}
	s.recordSagaTransition(ctx, sagaID, sagaReservationCreated, "success", fmt.Sprintf("Reserva %s creada en PostgreSQL", reservationId))

	notifBaseCtx := context.WithoutCancel(ctx)
	s.recordSagaTransition(notifBaseCtx, sagaID, sagaNotificationDispatching, "success", "Notificacion asincrona iniciada")
	go func() {
		notifCtx, cancel := context.WithTimeout(notifBaseCtx, 3*time.Second)
		defer cancel()
		_, err := Execute(s.notificationBreaker, func() (*pb.SendConfirmationResponse, error) {
			return s.notificationClient.SendConfirmation(notifCtx, &pb.SendConfirmationRequest{
				UserId:        req.UserId,
				ReservationId: reservationId,
				Tipo:          "CONFIRMACION",
				Email:         userEmail,
			})
		})
		if err != nil {
			log.Printf("[Reservas] saga_id=%s error al enviar notificacion: %v", sagaID, err)
			reservationsNotificationsTotal.WithLabelValues("failed").Inc()
			s.recordSagaTransition(context.Background(), sagaID, sagaCompletedWithNotificationFailure, "failed", status.Convert(err).Message())
		} else {
			log.Printf("[Reservas] saga_id=%s notificacion enviada exitosamente", sagaID)
			reservationsNotificationsTotal.WithLabelValues("success").Inc()
			s.recordSagaTransition(context.Background(), sagaID, sagaCompleted, "success", "Reserva confirmada y notificacion procesada")
		}
	}()

	reservationsCreatedTotal.Inc()
	return &pb.CreateReservationResponse{
		ReservationId: reservationId,
		Status:        "CONFIRMADA",
		MontoTotal:    montoTotal,
	}, nil
}

func main() {
	startMetricsServer(getEnv("METRICS_PORT", "9102"))
	tracerShutdown, err := initTracer(context.Background(), getEnv("OTEL_SERVICE_NAME", "reservation-service"))
	if err != nil {
		log.Printf("[Reservas] OpenTelemetry no pudo inicializarse: %v", err)
	} else {
		defer func() {
			if err := tracerShutdown(context.Background()); err != nil {
				log.Printf("[Reservas] Error cerrando OpenTelemetry: %v", err)
			}
		}()
	}

	dbHost := getEnv("RESERVAS_DB_HOST", "localhost")
	dbPort := getEnv("RESERVAS_DB_PORT", "5432")
	dbUser := getEnv("RESERVAS_DB_USER", "postgres")
	dbPassword := getEnv("RESERVAS_DB_PASSWORD", "postgres")
	dbName := getEnv("RESERVAS_DB_NAME", "reservas_service")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error inicializando conexion: %v", err)
	}

	for i := 0; i < 15; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Conexion exitosa a la base de datos de reservas")
			break
		}
		log.Printf("Esperando a la base de datos de reservas (%s)... intento %d/15", dbHost, i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("No se pudo conectar a la BD tras 15 intentos: %v", err)
	}
	defer db.Close()

	if err := initializeReservationSchema(db); err != nil {
		log.Fatalf("Error inicializando esquema de reservas y saga: %v", err)
	}

	userHost := getEnv("USER_SERVICE_HOST", "localhost:9090")
	userConn, err := grpc.Dial(userHost, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("Error conectando a Usuarios (%s): %v", userHost, err)
	}
	defer userConn.Close()

	invHost := getEnv("INVENTORY_SERVICE_HOST", "localhost:50053")
	invConn, err := grpc.Dial(invHost, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("Error conectando a Inventario (%s): %v", invHost, err)
	}
	defer invConn.Close()

	notifHost := getEnv("NOTIFICATION_SERVICE_HOST", "localhost:50051")
	notifConn, err := grpc.Dial(notifHost, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("Error conectando a Notificaciones (%s): %v", notifHost, err)
	}
	defer notifConn.Close()

	resServer := &reservationServer{
		userClient:         pb.NewUserServiceClient(userConn),
		inventoryClient:    pb.NewInventoryServiceClient(invConn),
		notificationClient: pb.NewNotificationServiceClient(notifConn),
		db:                 db,

		userBreaker:         NewCircuitBreaker("user", 5, 15*time.Second),
		inventoryBreaker:    NewCircuitBreaker("inventory", 5, 15*time.Second),
		notificationBreaker: NewCircuitBreaker("notification", 5, 15*time.Second),
	}

	port := getEnv("PORT", "50052")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("No se pudo escuchar en :%s: %v", port, err)
	}

	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterReservationServiceServer(srv, resServer)
	reflection.Register(srv)

	log.Printf("Servidor de Reservas escuchando en el puerto :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}
}
