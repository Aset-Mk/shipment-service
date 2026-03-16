package domain

// Status represents the current stage of a shipment in its lifecycle.
type Status string

const (
	StatusPending   Status = "pending"
	StatusPickedUp  Status = "picked_up"
	StatusInTransit Status = "in_transit"
	StatusDelivered Status = "delivered"
	StatusCancelled Status = "cancelled"
)

// allowedTransitions defines which status changes are considered valid.
// Terminal statuses (delivered, cancelled) have no outgoing transitions.
var allowedTransitions = map[Status][]Status{
	StatusPending:   {StatusPickedUp, StatusCancelled},
	StatusPickedUp:  {StatusInTransit, StatusCancelled},
	StatusInTransit: {StatusDelivered, StatusCancelled},
	StatusDelivered: {},
	StatusCancelled: {},
}

// IsValid reports whether the status value is one of the known statuses.
func (s Status) IsValid() bool {
	_, ok := allowedTransitions[s]
	return ok
}

// CanTransitionTo reports whether transitioning from s to next is allowed.
func (s Status) CanTransitionTo(next Status) bool {
	for _, allowed := range allowedTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

// IsTerminal reports whether the status has no further transitions.
func (s Status) IsTerminal() bool {
	return len(allowedTransitions[s]) == 0
}
