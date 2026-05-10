package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	_ "github.com/lib/pq"
	notifpb "github.com/darkkings19/mi-servicio/pb/notification"
	pb "github.com/darkkings19/mi-servicio/pb"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ─ Servidor de Reservas ─────────────────────────────────────────────
type reservationServer struct {
	pb.UnimplementedReservationServiceServer
	userClient         pb.UserServiceClient
	inventoryClient    pb.InventoryServiceClient
	notificationClient notifpb.NotificationServiceClient
	db                 *sql.DB
}

func (s *reservationServer) GetReservation(ctx context.Context, req *pb.GetReservationRequest) (*pb.GetReservationResponse, error) {
	return &pb.GetReservationResponse{
		Reservation: &pb.Reservation{
			ReservationId: req.ReservationId,
			Status:        "CONFIRMADA",
		},
	}, nil
}

func (s *reservationServer) ListReservations(ctx context.Context, req *pb.ListReservationsRequest) (*pb.ListReservationsResponse, error) {
	log.Println("[Reservas] Obteniendo lista de reservas desde la base de datos...")
	
	query := "SELECT reservation_id, user_id, hotel_id, room_type_id, status, monto_total FROM reservations ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[Reservas] Error leyendo base de datos: %v", err)
		return nil, status.Errorf(codes.Internal, "error al consultar la base de datos")
	}
	defer rows.Close()

	var list []*pb.Reservation
	for rows.Next() {
		var res pb.Reservation
		if err := rows.Scan(&res.ReservationId, &res.UserId, &res.HotelId, &res.RoomTypeId, &res.Status, &res.MontoTotal); err != nil {
			log.Printf("[Reservas] Error escaneando fila: %v", err)
			continue
		}
		list = append(list, &res)
	}

	return &pb.ListReservationsResponse{
		Reservations: list,
	}, nil
}

func (s *reservationServer) CreateReservation(ctx context.Context, req *pb.CreateReservationRequest) (*pb.CreateReservationResponse, error) {
	log.Printf("[Reservas] Iniciando creación de reserva para usuario: %s", req.UserId)

	// 1. Validar Usuario (MOCK LOCAL - Sin llamar al servicio real)
	log.Printf("[Reservas] (Mock) Validando usuario %s...", req.UserId)
	if req.UserId == "error_user" {
		return nil, status.Errorf(codes.NotFound, "usuario no encontrado")
	}

	// 2. Bloquear Inventario (MOCK LOCAL - Sin llamar al servicio real)
	log.Printf("[Reservas] (Mock) Solicitando bloqueo de inventario para %s...", req.RoomTypeId)
	if req.RoomTypeId == "agotado" {
		return nil, status.Errorf(codes.ResourceExhausted, "inventario agotado o rechazado")
	}

	// 3. Crear Reserva en DB (PostgreSQL)
	reservationId := "res-" + time.Now().Format("20060102150405")
	montoTotal := 150.50 // Monto simulado

	query := `INSERT INTO reservations (reservation_id, user_id, hotel_id, room_type_id, fecha_inicio, fecha_fin, status, monto_total) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, errDb := s.db.ExecContext(ctx, query, 
		reservationId, req.UserId, req.HotelId, req.RoomTypeId, 
		req.FechaInicio, req.FechaFin, "CONFIRMADA", montoTotal)
	
	if errDb != nil {
		log.Printf("[Reservas] Error al guardar en PostgreSQL: %v", errDb)
		return nil, status.Errorf(codes.Internal, "error interno al guardar la reserva en la base de datos")
	}

	log.Printf("[Reservas] Reserva %s asegurada en base de datos PostgreSQL con estado CONFIRMADA", reservationId)

	// 4. Enviar Notificación al servicio real (fire-and-forget — no bloquea la respuesta)
	go func() {
		notifCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := s.notificationClient.SendConfirmation(notifCtx, &notifpb.SendConfirmationRequest{
			UserId:        req.UserId,
			ReservationId: reservationId,
			Tipo:          "CONFIRMACION",
		})
		if err != nil {
			log.Printf("[Reservas] Error al enviar notificación: %v (la reserva ya está asegurada en DB)", err)
		} else {
			log.Printf("[Reservas] Notificación enviada exitosamente al servicio de notificaciones")
		}
	}()

	// 5. Retornar Respuesta al Cliente (API Gateway)
	return &pb.CreateReservationResponse{
		ReservationId: reservationId,
		Status:        "CONFIRMADA",
		MontoTotal:    montoTotal,
	}, nil
}

// ─ Main ─────────────────────────────────────────────
func main() {
	// --- Configuración Base de Datos PostgreSQL ---
	dbHost := getEnv("RESERVAS_DB_HOST", "localhost")
	dbPort := getEnv("RESERVAS_DB_PORT", "5432")
	dbUser := getEnv("RESERVAS_DB_USER", "postgres")
	dbPassword := getEnv("RESERVAS_DB_PASSWORD", "postgres")
	dbName := getEnv("RESERVAS_DB_NAME", "reservas_service")
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al preparar la conexión a PostgreSQL: %v", err)
	}
	defer db.Close()

	// Probar conexión a la BD
	if err = db.Ping(); err != nil {
		log.Printf("ADVERTENCIA: No se pudo conectar a PostgreSQL: %v", err)
		log.Println("Por favor verifica el usuario, contraseña y base de datos en la variable connStr.")
	} else {
		log.Println("Conexión a PostgreSQL establecida exitosamente.")
		
		// Crear tabla automáticamente si no existe
		createTableQuery := `
		CREATE TABLE IF NOT EXISTS reservations (
			reservation_id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL,
			hotel_id VARCHAR(50) NOT NULL,
			room_type_id VARCHAR(50) NOT NULL,
			fecha_inicio VARCHAR(20) NOT NULL,
			fecha_fin VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL,
			monto_total NUMERIC(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
		if _, err := db.Exec(createTableQuery); err != nil {
			log.Fatalf("Error creando tabla de reservaciones: %v", err)
		}
	}

	// --- Configuración de Clientes gRPC ---
	userConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("Error conectando a Usuarios: %v", err) }
	defer userConn.Close()

	invConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("Error conectando a Inventario: %v", err) }
	defer invConn.Close()

	notifHost := getEnv("NOTIFICATION_SERVICE_HOST", "localhost:50051")
	notifConn, err := grpc.Dial(notifHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("Error conectando a Notificaciones (%s): %v", notifHost, err) }
	defer notifConn.Close()

	resServer := &reservationServer{
		userClient:         pb.NewUserServiceClient(userConn),
		inventoryClient:    pb.NewInventoryServiceClient(invConn),
		notificationClient: notifpb.NewNotificationServiceClient(notifConn),
		db:                 db,
	}

	// --- Iniciar Servidor de Reservas ---
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("No se pudo escuchar en :50051: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterReservationServiceServer(srv, resServer)
	reflection.Register(srv)

	log.Println("Servidor de Reservas escuchando en el puerto :50051")
	log.Println("Esperando dependencias externas: Usuarios(50052), Inventario(50053), Notificaciones(...)")
	
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Error al servir Reservas: %v", err)
	}
}
