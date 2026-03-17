package usecase

import (
	"context"

	"github.com/Aset-Mk/shipment-service/internal/domain"
)

// CreateShipmentInput carries the data needed to create a new shipment.
type CreateShipmentInput struct {
	Reference     string
	Origin        string
	Destination   string
	Driver        domain.DriverInfo
	Unit          domain.UnitInfo
	Amount        float64
	DriverRevenue float64
}

// ShipmentUseCase defines the application-level operations on shipments.
// The transport layer depends on this interface, not on the concrete service.
type ShipmentUseCase interface {
	CreateShipment(ctx context.Context, input CreateShipmentInput) (*domain.Shipment, error)
	GetShipment(ctx context.Context, id string) (*domain.Shipment, error)
	AddEvent(ctx context.Context, shipmentID string, status domain.Status, note string) (*domain.ShipmentEvent, error)
	GetEvents(ctx context.Context, shipmentID string) ([]*domain.ShipmentEvent, error)
}
