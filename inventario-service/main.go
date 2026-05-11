package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	pb "github.com/darkkings19/inventario-service/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type inventoryServer struct {
	pb.UnimplementedInventoryServiceServer
	db *sql.DB
}

// --- Implementación de gRPC ---

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
		log.Printf("[Inventario] Error en búsqueda: %v", err)
		return nil, status.Errorf(codes.Internal, "error al consultar inventario")
	}
	defer rows.Close()

	var rooms []*pb.RoomTypeAvailability
	for rows.Next() {
		var r pb.RoomTypeAvailability
		if err := rows.Scan(&r.HotelId, &r.NombreHotel, &r.RoomTypeId, &r.PrecioNoche, &r.Capacidad, &r.StockDisponible, &r.RoomTypeName); err != nil {
			continue
		}
		rooms = append(rooms, &r)
	}

	return &pb.SearchResponse{Rooms: rooms}, nil
}

func (s *inventoryServer) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	var query string
	accionStr := fmt.Sprint(req.Accion)
	
	if accionStr == "BLOQUEAR" || accionStr == "0" {
		log.Printf("[Inventario] Bloqueando %d unidades de %s", req.Cantidad, req.RoomTypeId)
		query = "UPDATE room_types SET stock_total = stock_total - $1 WHERE id = $2 AND stock_total >= $1"
	} else if accionStr == "LIBERAR" || accionStr == "1" {
		log.Printf("[Inventario] Liberando %d unidades de %s", req.Cantidad, req.RoomTypeId)
		query = "UPDATE room_types SET stock_total = stock_total + $1 WHERE id = $2"
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "acción inválida: %s", accionStr)
	}

	res, err := s.db.ExecContext(ctx, query, req.Cantidad, req.RoomTypeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error al actualizar stock")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 && (accionStr == "BLOQUEAR" || accionStr == "0") {
		return nil, status.Errorf(codes.ResourceExhausted, "no hay stock suficiente")
	}

	return &pb.UpdateStockResponse{Status: rows > 0}, nil
}

// --- Inicialización y Seed ---

func (s *inventoryServer) initializeDB() {
	// Crear tablas
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

	// SEED: Insertar datos de prueba si las tablas están vacías
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM hotels").Scan(&count)
	if count == 0 {
		log.Println("[Inventario] Insertando datos de prueba (Seed)...")
		s.db.Exec("INSERT INTO hotels (id, nombre, ubicacion) VALUES ('h1', 'Hotel Continental', 'New York'), ('h2', 'Hotel California', 'California')")
		s.db.Exec("INSERT INTO room_types (id, hotel_id, nombre, precio_noche, capacidad, stock_total) VALUES ('rt1', 'h1', 'Suite Ejecutiva', 250.00, 2, 10), ('rt2', 'h1', 'Habitación Simple', 100.00, 1, 20), ('rt3', 'h2', 'Deluxe King', 180.00, 2, 5)")
	}
}

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "inventario_db")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error inicializando conexión: %v", err)
	}

	// Reintento de conexión (Ping)
	for i := 0; i < 15; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Conexión exitosa a la base de datos de inventario")
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

	s := grpc.NewServer()
	pb.RegisterInventoryServiceServer(s, server)
	reflection.Register(s)

	log.Printf("Servicio de Inventario escuchando en el puerto :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}
}
