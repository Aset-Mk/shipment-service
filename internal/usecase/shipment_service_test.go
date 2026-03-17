package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Aset-Mk/shipment-service/internal/domain"
)

// newTestService builds a service backed by in-memory repos with a
// deterministic id counter so tests are easy to reason about.
func newTestService() (ShipmentUseCase, *mockShipmentRepo, *mockEventRepo) {
	shipRepo := newMockShipmentRepo()
	eventRepo := &mockEventRepo{}
	counter := 0
	idGen := func() string {
		counter++
		return fmt.Sprintf("id-%d", counter)
	}
	svc := NewShipmentService(shipRepo, eventRepo, idGen)
	return svc, shipRepo, eventRepo
}

// --- CreateShipment ---

func TestCreateShipment_OK(t *testing.T) {
	svc, _, eventRepo := newTestService()
	ctx := context.Background()

	input := CreateShipmentInput{
		Reference:   "REF-001",
		Origin:      "Almaty",
		Destination: "Astana",
		Driver:      domain.DriverInfo{Name: "Ali", License: "AA1234"},
		Unit:        domain.UnitInfo{ID: "TRUCK-01", Type: "truck"},
		Amount:      1500.0,
		DriverRevenue: 300.0,
	}

	shipment, err := svc.CreateShipment(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if shipment.Status != domain.StatusPending {
		t.Errorf("expected status pending, got %s", shipment.Status)
	}
	if shipment.Reference != "REF-001" {
		t.Errorf("unexpected reference: %s", shipment.Reference)
	}

	// initial event must be recorded
	events, _ := eventRepo.FindByShipmentID(ctx, shipment.ID)
	if len(events) != 1 {
		t.Fatalf("expected 1 initial event, got %d", len(events))
	}
	if events[0].Status != domain.StatusPending {
		t.Errorf("initial event status should be pending, got %s", events[0].Status)
	}
}

func TestCreateShipment_DuplicateReference(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	input := CreateShipmentInput{
		Reference:   "REF-DUP",
		Origin:      "Almaty",
		Destination: "Astana",
	}

	if _, err := svc.CreateShipment(ctx, input); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err := svc.CreateShipment(ctx, input)
	if !errors.Is(err, domain.ErrDuplicateReference) {
		t.Errorf("expected ErrDuplicateReference, got %v", err)
	}
}

func TestCreateShipment_MissingFields(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	cases := []CreateShipmentInput{
		{Origin: "A", Destination: "B"},          // no reference
		{Reference: "R", Destination: "B"},        // no origin
		{Reference: "R", Origin: "A"},             // no destination
	}

	for _, input := range cases {
		_, err := svc.CreateShipment(ctx, input)
		if err == nil {
			t.Errorf("expected validation error for input %+v", input)
		}
	}
}

// --- AddEvent / status transitions ---

func TestAddEvent_ValidTransitions(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, CreateShipmentInput{
		Reference: "REF-002", Origin: "A", Destination: "B",
	})

	transitions := []domain.Status{
		domain.StatusPickedUp,
		domain.StatusInTransit,
		domain.StatusDelivered,
	}

	for _, next := range transitions {
		event, err := svc.AddEvent(ctx, shipment.ID, next, "")
		if err != nil {
			t.Fatalf("unexpected error on transition to %s: %v", next, err)
		}
		if event.Status != next {
			t.Errorf("event status mismatch: want %s, got %s", next, event.Status)
		}
		// verify the shipment's current status was updated
		updated, _ := svc.GetShipment(ctx, shipment.ID)
		if updated.Status != next {
			t.Errorf("shipment status not updated: want %s, got %s", next, updated.Status)
		}
	}
}

func TestAddEvent_InvalidTransition(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, CreateShipmentInput{
		Reference: "REF-003", Origin: "A", Destination: "B",
	})

	// pending → delivered is not allowed (must go through picked_up first)
	_, err := svc.AddEvent(ctx, shipment.ID, domain.StatusDelivered, "")
	if err == nil {
		t.Fatal("expected error for invalid transition, got nil")
	}

	var transErr *domain.ErrInvalidTransition
	if !errors.As(err, &transErr) {
		t.Errorf("expected ErrInvalidTransition, got %T: %v", err, err)
	}
}

func TestAddEvent_TransitionFromTerminalStatus(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, CreateShipmentInput{
		Reference: "REF-004", Origin: "A", Destination: "B",
	})

	// cancel the shipment
	if _, err := svc.AddEvent(ctx, shipment.ID, domain.StatusCancelled, "cancelled by client"); err != nil {
		t.Fatalf("unexpected error cancelling: %v", err)
	}

	// any further transition must be rejected
	_, err := svc.AddEvent(ctx, shipment.ID, domain.StatusPickedUp, "")
	if err == nil {
		t.Fatal("expected error transitioning from terminal status, got nil")
	}
}

func TestAddEvent_ShipmentNotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.AddEvent(ctx, "ghost-id", domain.StatusPickedUp, "")
	if !errors.Is(err, domain.ErrShipmentNotFound) {
		t.Errorf("expected ErrShipmentNotFound, got %v", err)
	}
}

// --- GetEvents ---

func TestGetEvents_FullHistory(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, CreateShipmentInput{
		Reference: "REF-005", Origin: "A", Destination: "B",
	})
	svc.AddEvent(ctx, shipment.ID, domain.StatusPickedUp, "driver arrived")
	svc.AddEvent(ctx, shipment.ID, domain.StatusInTransit, "on the road")

	events, err := svc.GetEvents(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// created → picked_up → in_transit = 3 events
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}
