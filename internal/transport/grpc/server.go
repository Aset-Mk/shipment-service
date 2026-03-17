package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Aset-Mk/shipment-service/internal/usecase"
	pb "github.com/Aset-Mk/shipment-service/gen/shipment"
)

// Server wraps the gRPC server and handles its lifecycle.
type Server struct {
	grpc *grpc.Server
	addr string
}

// NewServer creates a gRPC Server and registers all service handlers.
func NewServer(addr string, svc usecase.ShipmentUseCase) *Server {
	grpcSrv := grpc.NewServer()

	pb.RegisterShipmentServiceServer(grpcSrv, newHandler(svc))

	// reflection allows tools like grpcurl to discover the service at runtime
	reflection.Register(grpcSrv)

	return &Server{grpc: grpcSrv, addr: addr}
}

// Run starts listening and blocks until the process receives SIGINT or SIGTERM,
// then performs a graceful shutdown.
func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("gRPC server listening on %s", s.addr)
		if err := s.grpc.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		log.Printf("received signal %s, shutting down gracefully", sig)
		s.grpc.GracefulStop()
		return nil
	}
}

// Shutdown stops the server without waiting for ongoing RPCs to finish.
// Intended for use in tests or when a hard stop is needed.
func (s *Server) Shutdown(_ context.Context) {
	s.grpc.Stop()
}
