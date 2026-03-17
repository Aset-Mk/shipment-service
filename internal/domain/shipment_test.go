package domain

import (
	"errors"
	"testing"
	"time"
)

var testNow = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func newTestShipment() *Shipment {
	return NewShipment(
		"ship-1",
		"REF-TEST",
		"Almaty",
		"Astana",
		DriverInfo{Name: "Ali", License: "AA1234"},
		UnitInfo{ID: "TRUCK-01", Type: "truck"},
		1500.0,
		300.0,
		testNow,
	)
}

func TestNewShipment_InitialState(t *testing.T) {
	s := newTestShipment()

	if s.Status != StatusPending {
		t.Errorf("new shipment should start as pending, got %s", s.Status)
	}
	if s.ID != "ship-1" {
		t.Errorf("unexpected id: %s", s.ID)
	}
	if s.CreatedAt != testNow || s.UpdatedAt != testNow {
		t.Error("timestamps not set correctly on creation")
	}
}

func TestShipment_ApplyEvent_ValidChain(t *testing.T) {
	s := newTestShipment()
	now := testNow.Add(time.Hour)

	steps := []Status{StatusPickedUp, StatusInTransit, StatusDelivered}
	for _, next := range steps {
		if err := s.ApplyEvent(next, now); err != nil {
			t.Fatalf("unexpected error transitioning to %s: %v", next, err)
		}
		if s.Status != next {
			t.Errorf("after ApplyEvent(%s): status is %s", next, s.Status)
		}
		if s.UpdatedAt != now {
			t.Errorf("UpdatedAt not updated on transition to %s", next)
		}
	}
}

func TestShipment_ApplyEvent_CancelAtAnyNonTerminalStage(t *testing.T) {
	stages := []Status{StatusPending, StatusPickedUp, StatusInTransit}

	for _, stage := range stages {
		s := newTestShipment()
		// advance to the target stage
		path := transitionPathTo(stage)
		for _, st := range path {
			if err := s.ApplyEvent(st, testNow); err != nil {
				t.Fatalf("setup failed at %s: %v", st, err)
			}
		}
		// now cancel
		if err := s.ApplyEvent(StatusCancelled, testNow); err != nil {
			t.Errorf("expected cancellation from %s to succeed, got: %v", stage, err)
		}
	}
}

func TestShipment_ApplyEvent_InvalidTransition(t *testing.T) {
	s := newTestShipment()

	err := s.ApplyEvent(StatusDelivered, testNow)
	if err == nil {
		t.Fatal("expected error for invalid transition pending → delivered")
	}

	var transErr *ErrInvalidTransition
	if !errors.As(err, &transErr) {
		t.Errorf("expected ErrInvalidTransition, got %T: %v", err, err)
	}
	if transErr.From != StatusPending || transErr.To != StatusDelivered {
		t.Errorf("error fields mismatch: from=%s to=%s", transErr.From, transErr.To)
	}

	// status must not change on a rejected transition
	if s.Status != StatusPending {
		t.Errorf("status should remain pending after rejected transition, got %s", s.Status)
	}
}

func TestShipment_ApplyEvent_UnknownStatus(t *testing.T) {
	s := newTestShipment()

	if err := s.ApplyEvent(Status("flying"), testNow); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestShipment_ApplyEvent_TerminalStatusIsLocked(t *testing.T) {
	terminal := []Status{StatusDelivered, StatusCancelled}

	for _, term := range terminal {
		s := newTestShipment()
		for _, st := range transitionPathTo(term) {
			s.ApplyEvent(st, testNow) //nolint:errcheck
		}
		if s.Status != term {
			t.Fatalf("failed to reach terminal status %s", term)
		}

		for _, next := range []Status{StatusPending, StatusPickedUp, StatusInTransit, StatusDelivered, StatusCancelled} {
			if err := s.ApplyEvent(next, testNow); err == nil {
				t.Errorf("expected error transitioning from terminal %s to %s", term, next)
			}
		}
	}
}

func TestShipment_CanTransitionTo(t *testing.T) {
	s := newTestShipment() // status: pending

	if !s.CanTransitionTo(StatusPickedUp) {
		t.Error("pending should be able to transition to picked_up")
	}
	if s.CanTransitionTo(StatusDelivered) {
		t.Error("pending should not be able to transition to delivered")
	}
}

// transitionPathTo returns the sequence of ApplyEvent calls needed to
// bring a fresh pending shipment to the target status.
func transitionPathTo(target Status) []Status {
	switch target {
	case StatusPending:
		return nil
	case StatusPickedUp:
		return []Status{StatusPickedUp}
	case StatusInTransit:
		return []Status{StatusPickedUp, StatusInTransit}
	case StatusDelivered:
		return []Status{StatusPickedUp, StatusInTransit, StatusDelivered}
	case StatusCancelled:
		return []Status{StatusCancelled}
	default:
		return nil
	}
}
