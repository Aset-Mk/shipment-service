package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Aset-Mk/shipment-service/internal/domain"
	pb "github.com/Aset-Mk/shipment-service/gen/shipment"
)

// --- domain → proto ---

func shipmentToProto(s *domain.Shipment) *pb.Shipment {
	return &pb.Shipment{
		Id:            s.ID,
		Reference:     s.Reference,
		Origin:        s.Origin,
		Destination:   s.Destination,
		Status:        statusToProto(s.Status),
		Driver:        driverToProto(s.Driver),
		Unit:          unitToProto(s.Unit),
		Amount:        s.Amount,
		DriverRevenue: s.DriverRevenue,
		CreatedAt:     timestamppb.New(s.CreatedAt),
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
	}
}

func eventToProto(e *domain.ShipmentEvent) *pb.ShipmentEvent {
	return &pb.ShipmentEvent{
		Id:         e.ID,
		ShipmentId: e.ShipmentID,
		Status:     statusToProto(e.Status),
		Note:       e.Note,
		CreatedAt:  timestamppb.New(e.CreatedAt),
	}
}

func driverToProto(d domain.DriverInfo) *pb.DriverInfo {
	return &pb.DriverInfo{Name: d.Name, License: d.License}
}

func unitToProto(u domain.UnitInfo) *pb.UnitInfo {
	return &pb.UnitInfo{Id: u.ID, Type: u.Type}
}

func statusToProto(s domain.Status) pb.ShipmentStatus {
	switch s {
	case domain.StatusPending:
		return pb.ShipmentStatus_SHIPMENT_STATUS_PENDING
	case domain.StatusPickedUp:
		return pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP
	case domain.StatusInTransit:
		return pb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT
	case domain.StatusDelivered:
		return pb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED
	case domain.StatusCancelled:
		return pb.ShipmentStatus_SHIPMENT_STATUS_CANCELLED
	default:
		return pb.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED
	}
}

// --- proto → domain ---

func statusFromProto(s pb.ShipmentStatus) domain.Status {
	switch s {
	case pb.ShipmentStatus_SHIPMENT_STATUS_PENDING:
		return domain.StatusPending
	case pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP:
		return domain.StatusPickedUp
	case pb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT:
		return domain.StatusInTransit
	case pb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED:
		return domain.StatusDelivered
	case pb.ShipmentStatus_SHIPMENT_STATUS_CANCELLED:
		return domain.StatusCancelled
	default:
		return ""
	}
}

func driverFromProto(d *pb.DriverInfo) domain.DriverInfo {
	if d == nil {
		return domain.DriverInfo{}
	}
	return domain.DriverInfo{Name: d.Name, License: d.License}
}

func unitFromProto(u *pb.UnitInfo) domain.UnitInfo {
	if u == nil {
		return domain.UnitInfo{}
	}
	return domain.UnitInfo{ID: u.Id, Type: u.Type}
}
