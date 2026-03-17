package usecase

import (
	"context"
	"sync"

	"github.com/Aset-Mk/shipment-service/internal/domain"
)

// --- in-memory ShipmentRepository ---

type mockShipmentRepo struct {
	mu        sync.RWMutex
	byID      map[string]*domain.Shipment
	byRef     map[string]*domain.Shipment
}

func newMockShipmentRepo() *mockShipmentRepo {
	return &mockShipmentRepo{
		byID:  make(map[string]*domain.Shipment),
		byRef: make(map[string]*domain.Shipment),
	}
}

func (r *mockShipmentRepo) Save(_ context.Context, s *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *s
	r.byID[s.ID] = &cp
	r.byRef[s.Reference] = &cp
	return nil
}

func (r *mockShipmentRepo) Update(_ context.Context, s *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *s
	r.byID[s.ID] = &cp
	r.byRef[s.Reference] = &cp
	return nil
}

func (r *mockShipmentRepo) FindByID(_ context.Context, id string) (*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrShipmentNotFound
	}
	cp := *s
	return &cp, nil
}

func (r *mockShipmentRepo) FindByReference(_ context.Context, ref string) (*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byRef[ref]
	if !ok {
		return nil, domain.ErrShipmentNotFound
	}
	cp := *s
	return &cp, nil
}

// --- in-memory EventRepository ---

type mockEventRepo struct {
	mu     sync.RWMutex
	events []*domain.ShipmentEvent
}

func (r *mockEventRepo) Save(_ context.Context, e *domain.ShipmentEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *e
	r.events = append(r.events, &cp)
	return nil
}

func (r *mockEventRepo) FindByShipmentID(_ context.Context, id string) ([]*domain.ShipmentEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.ShipmentEvent
	for _, e := range r.events {
		if e.ShipmentID == id {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}
