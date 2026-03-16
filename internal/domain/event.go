package domain

import "time"

// ShipmentEvent records a single status change in a shipment's lifecycle.
// Events are immutable once created — the history is append-only.
type ShipmentEvent struct {
	ID         string
	ShipmentID string
	Status     Status
	Note       string
	CreatedAt  time.Time
}
