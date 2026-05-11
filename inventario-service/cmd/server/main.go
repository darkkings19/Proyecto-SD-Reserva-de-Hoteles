package main

import (
	"fmt"
	"log"
	"net"

	"github.com/darkkings19/inventario-service/internal/config"
	"github.com/darkkings19/inventario-service/internal/db"
	"github.com/darkkings19/inventario-service/internal/repository"
	"github.com/darkkings19/inventario-service/internal/service"
	handler "github.com/darkkings19/inventario-service/internal/transport/grpc"
	pb "github.com/darkkings19/inventario-service/proto/gen"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load .env file if present (ignored in Docker where env vars are injected)
	_ = godotenv.Load()

	cfg := config.Load()

	log.Printf("Connecting to database %s@%s:%s/%s ...", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	database, err := db.ConnectDB(cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	repo := repository.NewPostgresInventoryRepository(database)
	svc := service.NewInventoryService(repo)
	inventoryHandler := handler.NewInventoryHandler(svc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterInventoryServiceServer(grpcServer, inventoryHandler)

	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	log.Printf("Inventario-Service gRPC server starting on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
