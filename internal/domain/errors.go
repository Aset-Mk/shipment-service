package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrShipmentNotFound is returned when a shipment with the given id does not exist.
	ErrShipmentNotFound = errors.New("shipment not found")

	// ErrDuplicateReference is returned when a shipment with the same reference already exists.
	ErrDuplicateReference = errors.New("shipment with this reference already exists")

	// ErrInvalidStatus is returned when an unknown status value is provided.
	ErrInvalidStatus = errors.New("invalid shipment status")
)

// ErrInvalidTransition is returned when a requested status change violates the lifecycle rules.
type ErrInvalidTransition struct {
	From Status
	To   Status
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("transition from %q to %q is not allowed", e.From, e.To)
}
