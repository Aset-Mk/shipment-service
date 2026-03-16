package domain

import "time"

// DriverInfo holds identifying information about the driver assigned to a shipment.
type DriverInfo struct {
	Name    string
	License string
}

// UnitInfo describes the transport unit (truck, van, etc.) carrying the shipment.
type UnitInfo struct {
	ID   string
	Type string
}

// Shipment is the central aggregate of the domain.
// It owns its status and enforces all lifecycle rules.
type Shipment struct {
	ID            string
	Reference     string
	Origin        string
	Destination   string
	Status        Status
	Driver        DriverInfo
	Unit          UnitInfo
	Amount        float64
	DriverRevenue float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewShipment creates a new Shipment in the initial pending state.
// It does not persist anything — persistence is the caller's responsibility.
func NewShipment(
	id, reference, origin, destination string,
	driver DriverInfo,
	unit UnitInfo,
	amount, driverRevenue float64,
	now time.Time,
) *Shipment {
	return &Shipment{
		ID:            id,
		Reference:     reference,
		Origin:        origin,
		Destination:   destination,
		Status:        StatusPending,
		Driver:        driver,
		Unit:          unit,
		Amount:        amount,
		DriverRevenue: driverRevenue,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// CanTransitionTo reports whether the shipment may move to the given status.
func (s *Shipment) CanTransitionTo(next Status) bool {
	return s.Status.CanTransitionTo(next)
}

// ApplyEvent attempts to advance the shipment to the requested status.
// It returns an ErrInvalidTransition when the transition is not permitted,
// or ErrInvalidStatus when the status value itself is unknown.
func (s *Shipment) ApplyEvent(next Status, now time.Time) error {
	if !next.IsValid() {
		return ErrInvalidStatus
	}
	if !s.CanTransitionTo(next) {
		return &ErrInvalidTransition{From: s.Status, To: next}
	}
	s.Status = next
	s.UpdatedAt = now
	return nil
}
