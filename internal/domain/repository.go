package domain

import "context"

// ShipmentRepository defines persistence operations for Shipment aggregates.
// Implementations live in the infrastructure layer and are injected at startup.
type ShipmentRepository interface {
	Save(ctx context.Context, shipment *Shipment) error
	Update(ctx context.Context, shipment *Shipment) error
	FindByID(ctx context.Context, id string) (*Shipment, error)
	FindByReference(ctx context.Context, reference string) (*Shipment, error)
}

// EventRepository defines persistence operations for ShipmentEvent records.
type EventRepository interface {
	Save(ctx context.Context, event *ShipmentEvent) error
	FindByShipmentID(ctx context.Context, shipmentID string) ([]*ShipmentEvent, error)
}
