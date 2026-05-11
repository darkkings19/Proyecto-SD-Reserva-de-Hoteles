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
	notificationClient pb.NotificationServiceClient
	db                 *sql.DB
}

func (s *reservationServer) GetReservation(ctx context.Context, req *pb.GetReservationRequest) (*pb.GetReservationResponse, error) {
	log.Printf("[Reservas] Buscando reserva %s...", req.ReservationId)
	
	query := "SELECT reservation_id, user_id, hotel_id, room_type_id, status, monto_total, fecha_inicio, fecha_fin FROM reservations WHERE reservation_id = $1"
	var res pb.Reservation
	err := s.db.QueryRowContext(ctx, query, req.ReservationId).Scan(
		&res.ReservationId, &res.UserId, &res.HotelId, &res.RoomTypeId, 
		&res.Status, &res.MontoTotal, &res.FechaInicio, &res.FechaFin)
	
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "reserva no encontrada")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "error en base de datos: %v", err)
	}

	return &pb.GetReservationResponse{Reservation: &res}, nil
}

func (s *reservationServer) ListReservations(ctx context.Context, req *pb.ListReservationsRequest) (*pb.ListReservationsResponse, error) {
	log.Println("[Reservas] Obteniendo lista de reservas...")
	
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

	// 1. Validar Usuario Real en user-service
	log.Printf("[Reservas] Validando usuario %s en user-service...", req.UserId)
	userRes, errUser := s.userClient.GetUser(ctx, &pb.GetUserRequest{Id: req.UserId})
	if errUser != nil {
		log.Printf("[Reservas] Error al validar usuario: %v", errUser)
		return nil, status.Errorf(codes.Unauthenticated, "No se pudo validar el usuario: %v", status.Convert(errUser).Message())
	}
	log.Printf("[Reservas] Usuario validado: %s", userRes.User.Nombre)

	// 2. Bloquear Inventario REAL en el Servicio de Inventario
	log.Printf("[Reservas] Solicitando bloqueo de inventario para %s...", req.RoomTypeId)
	invRes, errInv := s.inventoryClient.UpdateStock(ctx, &pb.UpdateStockRequest{
		RoomTypeId: req.RoomTypeId,
		Cantidad:   1,
		Accion:     "BLOQUEAR",
	})
	
	if errInv != nil {
		log.Printf("[Reservas] Error al bloquear inventario: %v", errInv)
		return nil, status.Errorf(codes.ResourceExhausted, "No se pudo asegurar la habitación: %v", status.Convert(errInv).Message())
	}

	if !invRes.Status {
		return nil, status.Errorf(codes.ResourceExhausted, "No hay stock disponible")
	}
	log.Printf("[Reservas] Inventario bloqueado exitosamente para %s", req.RoomTypeId)

	// 3. Crear Reserva en DB (PostgreSQL)
	reservationId := "res-" + time.Now().Format("20060102150405")
	montoTotal := 150.50 // Monto simulado (debería venir del inventario o catálogo)

	query := `INSERT INTO reservations (reservation_id, user_id, hotel_id, room_type_id, fecha_inicio, fecha_fin, status, monto_total) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, errDb := s.db.ExecContext(ctx, query, 
		reservationId, req.UserId, req.HotelId, req.RoomTypeId, 
		req.FechaInicio, req.FechaFin, "CONFIRMADA", montoTotal)
	
	if errDb != nil {
		log.Printf("[Reservas] Error al guardar en PostgreSQL: %v", errDb)
		// Compensación: Liberar inventario si falló el guardado en DB
		go func() {
			_, _ = s.inventoryClient.UpdateStock(context.Background(), &pb.UpdateStockRequest{
				RoomTypeId: req.RoomTypeId,
				Cantidad:   1,
				Accion:     "LIBERAR",
			})
		}()
		return nil, status.Errorf(codes.Internal, "error interno al guardar la reserva")
	}

	log.Printf("[Reservas] Reserva %s creada en PostgreSQL", reservationId)

	// 4. Enviar Notificación al servicio real (fire-and-forget)
	go func() {
		notifCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := s.notificationClient.SendConfirmation(notifCtx, &pb.SendConfirmationRequest{
			UserId:        req.UserId,
			ReservationId: reservationId,
			Tipo:          "CONFIRMACION",
		})
		if err != nil {
			log.Printf("[Reservas] Error al enviar notificación: %v", err)
		} else {
			log.Printf("[Reservas] Notificación enviada exitosamente")
		}
	}()

	// 5. Retornar Respuesta
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
		log.Fatalf("Error inicializando conexión: %v", err)
	}

	// Reintento de conexión (Ping)
	for i := 0; i < 15; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("Conexión exitosa a la base de datos de reservas")
			break
		}
		log.Printf("Esperando a la base de datos de reservas (%s)... intento %d/15", dbHost, i+1)
		time.Sleep(3 * time.Second)
	}
	
	if err != nil { 
		log.Fatalf("No se pudo conectar a la BD tras 15 intentos: %v", err) 
	}
	defer db.Close()

	// Crear tabla automáticamente
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
		log.Fatalf("Error creando tabla: %v", err)
	}

	// --- Configuración de Clientes gRPC ---
	userHost := getEnv("USER_SERVICE_HOST", "localhost:9090")
	userConn, err := grpc.Dial(userHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("Error conectando a Usuarios (%s): %v", userHost, err) }
	defer userConn.Close()

	invHost := getEnv("INVENTORY_SERVICE_HOST", "localhost:50053")
	invConn, err := grpc.Dial(invHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("Error conectando a Inventario (%s): %v", invHost, err) }
	defer invConn.Close()

	notifHost := getEnv("NOTIFICATION_SERVICE_HOST", "localhost:50051")
	notifConn, err := grpc.Dial(notifHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("Error conectando a Notificaciones (%s): %v", notifHost, err) }
	defer notifConn.Close()

	resServer := &reservationServer{
		userClient:         pb.NewUserServiceClient(userConn),
		inventoryClient:    pb.NewInventoryServiceClient(invConn),
		notificationClient: pb.NewNotificationServiceClient(notifConn),
		db:                 db,
	}

	// --- Iniciar Servidor ---
	port := getEnv("PORT", "50052")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("No se pudo escuchar en :%s: %v", port, err)
	}

	srv := grpc.NewServer()
	pb.RegisterReservationServiceServer(srv, resServer)
	reflection.Register(srv)

	log.Printf("Servidor de Reservas escuchando en el puerto :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}
}
