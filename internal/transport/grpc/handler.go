package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Aset-Mk/shipment-service/internal/domain"
	"github.com/Aset-Mk/shipment-service/internal/usecase"
	pb "github.com/Aset-Mk/shipment-service/gen/shipment"
)

// handler implements pb.ShipmentServiceServer by delegating to the use-case layer.
type handler struct {
	pb.UnimplementedShipmentServiceServer
	svc usecase.ShipmentUseCase
}

func newHandler(svc usecase.ShipmentUseCase) *handler {
	return &handler{svc: svc}
}

func (h *handler) CreateShipment(ctx context.Context, req *pb.CreateShipmentRequest) (*pb.CreateShipmentResponse, error) {
	input := usecase.CreateShipmentInput{
		Reference:     req.Reference,
		Origin:        req.Origin,
		Destination:   req.Destination,
		Driver:        driverFromProto(req.Driver),
		Unit:          unitFromProto(req.Unit),
		Amount:        req.Amount,
		DriverRevenue: req.DriverRevenue,
	}

	shipment, err := h.svc.CreateShipment(ctx, input)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateShipmentResponse{Shipment: shipmentToProto(shipment)}, nil
}

func (h *handler) GetShipment(ctx context.Context, req *pb.GetShipmentRequest) (*pb.GetShipmentResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	shipment, err := h.svc.GetShipment(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetShipmentResponse{Shipment: shipmentToProto(shipment)}, nil
}

func (h *handler) AddEvent(ctx context.Context, req *pb.AddEventRequest) (*pb.AddEventResponse, error) {
	if req.ShipmentId == "" {
		return nil, status.Error(codes.InvalidArgument, "shipment_id is required")
	}
	if req.Status == pb.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	domainStatus := statusFromProto(req.Status)
	event, err := h.svc.AddEvent(ctx, req.ShipmentId, domainStatus, req.Note)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.AddEventResponse{Event: eventToProto(event)}, nil
}

func (h *handler) GetEvents(ctx context.Context, req *pb.GetEventsRequest) (*pb.GetEventsResponse, error) {
	if req.ShipmentId == "" {
		return nil, status.Error(codes.InvalidArgument, "shipment_id is required")
	}

	events, err := h.svc.GetEvents(ctx, req.ShipmentId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	protoEvents := make([]*pb.ShipmentEvent, len(events))
	for i, e := range events {
		protoEvents[i] = eventToProto(e)
	}
	return &pb.GetEventsResponse{Events: protoEvents}, nil
}

// toGRPCError maps domain errors to appropriate gRPC status codes.
func toGRPCError(err error) error {
	if errors.Is(err, domain.ErrShipmentNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	if errors.Is(err, domain.ErrDuplicateReference) {
		return status.Error(codes.AlreadyExists, err.Error())
	}
	if errors.Is(err, domain.ErrInvalidStatus) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	var transErr *domain.ErrInvalidTransition
	if errors.As(err, &transErr) {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
