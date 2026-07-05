package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "github.com/darkkings19/inventario-service/proto/gen"
	_ "github.com/lib/pq"
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
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	inventorySearchTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "inventory_search_total",
		Help: "Busquedas de habitaciones ejecutadas correctamente.",
	})
	inventorySearchEmptyTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "inventory_search_empty_total",
		Help: "Busquedas de habitaciones sin resultados.",
	})
	inventoryStockUpdatesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "inventory_stock_updates_total",
		Help: "Actualizaciones de stock por accion.",
	}, []string{"action"})
	inventoryErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "inventory_errors_total",
		Help: "Errores del dominio de inventario.",
	}, []string{"operation", "error_type"})
	inventoryAvailableRooms = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "inventory_available_rooms",
		Help: "Suma actual del stock disponible en inventario.",
	}, func() float64 {
		if inventoryDB == nil {
			return 0
		}
		var total float64
		if err := inventoryDB.QueryRow("SELECT COALESCE(SUM(stock_total), 0) FROM room_types").Scan(&total); err != nil {
			log.Printf("[Inventario] Error leyendo stock disponible para metrica: %v", err)
			return 0
		}
		return total
	})
)

var inventoryDB *sql.DB

func init() {
	prometheus.MustRegister(
		inventorySearchTotal,
		inventorySearchEmptyTotal,
		inventoryStockUpdatesTotal,
		inventoryErrorsTotal,
		inventoryAvailableRooms,
	)
	inventoryStockUpdatesTotal.WithLabelValues("bloquear").Add(0)
	inventoryStockUpdatesTotal.WithLabelValues("liberar").Add(0)
	inventoryErrorsTotal.WithLabelValues("search", "database").Add(0)
	inventoryErrorsTotal.WithLabelValues("search", "row_scan").Add(0)
	inventoryErrorsTotal.WithLabelValues("update_stock", "invalid_action").Add(0)
	inventoryErrorsTotal.WithLabelValues("update_stock", "database").Add(0)
	inventoryErrorsTotal.WithLabelValues("update_stock", "insufficient_stock").Add(0)
	inventoryErrorsTotal.WithLabelValues("update_stock", "room_type_not_found").Add(0)
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
		log.Printf("[Inventario] Metricas Prometheus en :%s/metrics", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Printf("[Inventario] Error en servidor de metricas: %v", err)
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

type inventoryServer struct {
	pb.UnimplementedInventoryServiceServer
	db *sql.DB
}

func (s *inventoryServer) SearchAvailableRooms(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	log.Printf("[Inventario] Buscando habitaciones en: %s", req.Ubicacion)

	query := `
		SELECT h.id, h.nombre, rt.id, rt.precio_noche, rt.capacidad, rt.stock_total, rt.nombre
		FROM room_types rt
		JOIN hotels h ON h.id = rt.hotel_id
		WHERE rt.stock_total > 0
	`
	args := []interface{}{}
	argCount := 1

	if req.Ubicacion != "" {
		query += fmt.Sprintf(" AND h.ubicacion ILIKE $%d", argCount)
		args = append(args, "%"+req.Ubicacion+"%")
		argCount++
	}
	if req.PrecioMax > 0 {
		query += fmt.Sprintf(" AND rt.precio_noche <= $%d", argCount)
		args = append(args, req.PrecioMax)
		argCount++
	}
	if req.Capacidad > 0 {
		query += fmt.Sprintf(" AND rt.capacidad >= $%d", argCount)
		args = append(args, req.Capacidad)
		argCount++
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("[Inventario] Error en busqueda: %v", err)
		inventoryErrorsTotal.WithLabelValues("search", "database").Inc()
		return nil, status.Errorf(codes.Internal, "error al consultar inventario")
	}
	defer rows.Close()

	var rooms []*pb.RoomTypeAvailability
	for rows.Next() {
		var r pb.RoomTypeAvailability
		if err := rows.Scan(&r.HotelId, &r.NombreHotel, &r.RoomTypeId, &r.PrecioNoche, &r.Capacidad, &r.StockDisponible, &r.RoomTypeName); err != nil {
			inventoryErrorsTotal.WithLabelValues("search", "row_scan").Inc()
			continue
		}
		rooms = append(rooms, &r)
	}

	inventorySearchTotal.Inc()
	if len(rooms) == 0 {
		inventorySearchEmptyTotal.Inc()
	}
	return &pb.SearchResponse{Rooms: rooms}, nil
}

func (s *inventoryServer) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	var query string
	actionLabel := ""
	accionStr := fmt.Sprint(req.Accion)

	if accionStr == "BLOQUEAR" || accionStr == "0" {
		actionLabel = "bloquear"
		log.Printf("[Inventario] Bloqueando %d unidades de %s", req.Cantidad, req.RoomTypeId)
		query = "UPDATE room_types SET stock_total = stock_total - $1 WHERE id = $2 AND stock_total >= $1"
	} else if accionStr == "LIBERAR" || accionStr == "1" {
		actionLabel = "liberar"
		log.Printf("[Inventario] Liberando %d unidades de %s", req.Cantidad, req.RoomTypeId)
		query = "UPDATE room_types SET stock_total = stock_total + $1 WHERE id = $2"
	} else {
		inventoryErrorsTotal.WithLabelValues("update_stock", "invalid_action").Inc()
		return nil, status.Errorf(codes.InvalidArgument, "accion invalida: %s", accionStr)
	}

	res, err := s.db.ExecContext(ctx, query, req.Cantidad, req.RoomTypeId)
	if err != nil {
		inventoryErrorsTotal.WithLabelValues("update_stock", "database").Inc()
		return nil, status.Errorf(codes.Internal, "error al actualizar stock")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 && actionLabel == "bloquear" {
		inventoryErrorsTotal.WithLabelValues("update_stock", "insufficient_stock").Inc()
		return nil, status.Errorf(codes.ResourceExhausted, "no hay stock suficiente")
	}
	if rows == 0 {
		inventoryErrorsTotal.WithLabelValues("update_stock", "room_type_not_found").Inc()
	}
	if rows > 0 {
		inventoryStockUpdatesTotal.WithLabelValues(actionLabel).Add(float64(req.Cantidad))
	}

	return &pb.UpdateStockResponse{Status: rows > 0}, nil
}

func (s *inventoryServer) initializeDB() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS hotels (
			id VARCHAR(50) PRIMARY KEY,
			nombre VARCHAR(100) NOT NULL,
			ubicacion VARCHAR(100) NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS room_types (
			id VARCHAR(50) PRIMARY KEY,
			hotel_id VARCHAR(50) REFERENCES hotels(id),
			nombre VARCHAR(50) NOT NULL,
			precio_noche NUMERIC(10, 2) NOT NULL,
			capacidad INT NOT NULL,
			stock_total INT NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			log.Fatalf("Error creando tablas: %v", err)
		}
	}

	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM hotels").Scan(&count)
	if count == 0 {
		log.Println("[Inventario] Insertando datos de prueba")
		s.db.Exec("INSERT INTO hotels (id, nombre, ubicacion) VALUES ('h1', 'Hotel Continental', 'New York'), ('h2', 'Hotel California', 'California')")
		s.db.Exec("INSERT INTO room_types (id, hotel_id, nombre, precio_noche, capacidad, stock_total) VALUES ('rt1', 'h1', 'Suite Ejecutiva', 250.00, 2, 10), ('rt2', 'h1', 'Habitacion Simple', 100.00, 1, 20), ('rt3', 'h2', 'Deluxe King', 180.00, 2, 5)")
	}
}

func main() {
	startMetricsServer(getEnv("METRICS_PORT", "9103"))
	tracerShutdown, err := initTracer(context.Background(), getEnv("OTEL_SERVICE_NAME", "inventory-service"))
	if err != nil {
		log.Printf("[Inventario] OpenTelemetry no pudo inicializarse: %v", err)
	} else {
		defer func() {
			if err := tracerShutdown(context.Background()); err != nil {
				log.Printf("[Inventario] Error cerrando OpenTelemetry: %v", err)
			}
		}()
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "inventario_db")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error inicializando conexion: %v", err)
	}
	inventoryDB = db

	for i := 0; i < 15; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Conexion exitosa a la base de datos de inventario")
			break
		}
		log.Printf("Esperando a la base de datos de inventario (%s)... intento %d/15", dbHost, i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("No se pudo conectar a la BD de inventario tras 15 intentos: %v", err)
	}

	server := &inventoryServer{db: db}
	server.initializeDB()

	port := getEnv("PORT", "50053")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}

	s := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterInventoryServiceServer(s, server)
	reflection.Register(s)

	log.Printf("Servicio de Inventario escuchando en el puerto :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}
}
