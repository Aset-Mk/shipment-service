package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Aset-Mk/shipment-service/internal/config"
	"github.com/Aset-Mk/shipment-service/internal/infrastructure/postgres"
	grpcserver "github.com/Aset-Mk/shipment-service/internal/transport/grpc"
	"github.com/Aset-Mk/shipment-service/internal/usecase"

	"github.com/google/uuid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Printf("database ping failed: %v", err)
		os.Exit(1)
	}
	log.Println("connected to database")

	// wire dependencies
	shipmentRepo := postgres.NewShipmentRepo(db)
	eventRepo := postgres.NewEventRepo(db)
	svc := usecase.NewShipmentService(shipmentRepo, eventRepo, uuid.NewString)

	srv := grpcserver.NewServer(cfg.GRPCAddr, svc)
	if err := srv.Run(); err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}
