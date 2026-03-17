package domain

import "testing"

func TestStatus_IsValid(t *testing.T) {
	valid := []Status{StatusPending, StatusPickedUp, StatusInTransit, StatusDelivered, StatusCancelled}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %q to be valid", s)
		}
	}

	if Status("garbage").IsValid() {
		t.Error("expected unknown status to be invalid")
	}
	if Status("").IsValid() {
		t.Error("expected empty status to be invalid")
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	terminal := []Status{StatusDelivered, StatusCancelled}
	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Errorf("expected %q to be terminal", s)
		}
	}

	nonTerminal := []Status{StatusPending, StatusPickedUp, StatusInTransit}
	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Errorf("expected %q to be non-terminal", s)
		}
	}
}

func TestStatus_CanTransitionTo(t *testing.T) {
	cases := []struct {
		from    Status
		to      Status
		allowed bool
	}{
		{StatusPending, StatusPickedUp, true},
		{StatusPending, StatusCancelled, true},
		{StatusPending, StatusInTransit, false},
		{StatusPending, StatusDelivered, false},

		{StatusPickedUp, StatusInTransit, true},
		{StatusPickedUp, StatusCancelled, true},
		{StatusPickedUp, StatusPending, false},
		{StatusPickedUp, StatusDelivered, false},

		{StatusInTransit, StatusDelivered, true},
		{StatusInTransit, StatusCancelled, true},
		{StatusInTransit, StatusPending, false},
		{StatusInTransit, StatusPickedUp, false},

		{StatusDelivered, StatusCancelled, false},
		{StatusDelivered, StatusPickedUp, false},
		{StatusCancelled, StatusPending, false},
		{StatusCancelled, StatusPickedUp, false},
	}

	for _, tc := range cases {
		got := tc.from.CanTransitionTo(tc.to)
		if got != tc.allowed {
			t.Errorf("(%s → %s): expected allowed=%v, got %v", tc.from, tc.to, tc.allowed, got)
		}
	}
}
