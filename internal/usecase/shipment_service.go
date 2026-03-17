package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Aset-Mk/shipment-service/internal/domain"
)

// IDGenerator is a function that returns a new unique identifier.
// Injected so the service stays testable without relying on a global uuid call.
type IDGenerator func() string

// shipmentService is the concrete implementation of ShipmentUseCase.
type shipmentService struct {
	shipments domain.ShipmentRepository
	events    domain.EventRepository
	newID     IDGenerator
	now       func() time.Time
}

// NewShipmentService constructs a shipmentService with the provided dependencies.
func NewShipmentService(
	shipments domain.ShipmentRepository,
	events domain.EventRepository,
	newID IDGenerator,
) ShipmentUseCase {
	return &shipmentService{
		shipments: shipments,
		events:    events,
		newID:     newID,
		now:       time.Now,
	}
}

// CreateShipment validates the input, creates a Shipment aggregate in the
// pending state, persists it together with the initial status event.
func (s *shipmentService) CreateShipment(ctx context.Context, input CreateShipmentInput) (*domain.Shipment, error) {
	if input.Reference == "" {
		return nil, fmt.Errorf("reference is required")
	}
	if input.Origin == "" || input.Destination == "" {
		return nil, fmt.Errorf("origin and destination are required")
	}

	existing, err := s.shipments.FindByReference(ctx, input.Reference)
	if err != nil && err != domain.ErrShipmentNotFound {
		return nil, fmt.Errorf("checking reference uniqueness: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrDuplicateReference
	}

	now := s.now()
	shipment := domain.NewShipment(
		s.newID(),
		input.Reference,
		input.Origin,
		input.Destination,
		input.Driver,
		input.Unit,
		input.Amount,
		input.DriverRevenue,
		now,
	)

	if err := s.shipments.Save(ctx, shipment); err != nil {
		return nil, fmt.Errorf("saving shipment: %w", err)
	}

	event := &domain.ShipmentEvent{
		ID:         s.newID(),
		ShipmentID: shipment.ID,
		Status:     domain.StatusPending,
		Note:       "shipment created",
		CreatedAt:  now,
	}
	if err := s.events.Save(ctx, event); err != nil {
		return nil, fmt.Errorf("saving initial event: %w", err)
	}

	return shipment, nil
}

// GetShipment returns the shipment by id, or ErrShipmentNotFound.
func (s *shipmentService) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	shipment, err := s.shipments.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

// AddEvent attempts to transition the shipment to the requested status.
// The domain aggregate enforces all transition rules; this method only
// orchestrates loading, mutating, and persisting.
func (s *shipmentService) AddEvent(ctx context.Context, shipmentID string, status domain.Status, note string) (*domain.ShipmentEvent, error) {
	shipment, err := s.shipments.FindByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	if err := shipment.ApplyEvent(status, now); err != nil {
		return nil, err
	}

	if err := s.shipments.Update(ctx, shipment); err != nil {
		return nil, fmt.Errorf("updating shipment: %w", err)
	}

	event := &domain.ShipmentEvent{
		ID:         s.newID(),
		ShipmentID: shipmentID,
		Status:     status,
		Note:       note,
		CreatedAt:  now,
	}
	if err := s.events.Save(ctx, event); err != nil {
		return nil, fmt.Errorf("saving event: %w", err)
	}

	return event, nil
}

// GetEvents returns the full status history for a shipment in chronological order.
func (s *shipmentService) GetEvents(ctx context.Context, shipmentID string) ([]*domain.ShipmentEvent, error) {
	if _, err := s.shipments.FindByID(ctx, shipmentID); err != nil {
		return nil, err
	}

	events, err := s.events.FindByShipmentID(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("fetching events: %w", err)
	}
	return events, nil
}
