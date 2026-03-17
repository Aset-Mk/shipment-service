package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Aset-Mk/shipment-service/internal/domain"
)

type eventRepo struct {
	db *pgxpool.Pool
}

// NewEventRepo creates a Postgres-backed EventRepository.
func NewEventRepo(db *pgxpool.Pool) domain.EventRepository {
	return &eventRepo{db: db}
}

const saveEventSQL = `
INSERT INTO shipment_events (id, shipment_id, status, note, created_at)
VALUES ($1, $2, $3, $4, $5)`

func (r *eventRepo) Save(ctx context.Context, e *domain.ShipmentEvent) error {
	_, err := r.db.Exec(ctx, saveEventSQL,
		e.ID, e.ShipmentID, string(e.Status), e.Note, e.CreatedAt,
	)
	return err
}

const findEventsByShipmentSQL = `
SELECT id, shipment_id, status, note, created_at
FROM shipment_events
WHERE shipment_id = $1
ORDER BY created_at ASC`

func (r *eventRepo) FindByShipmentID(ctx context.Context, shipmentID string) ([]*domain.ShipmentEvent, error) {
	rows, err := r.db.Query(ctx, findEventsByShipmentSQL, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.ShipmentEvent
	for rows.Next() {
		var e domain.ShipmentEvent
		var status string
		if err := rows.Scan(&e.ID, &e.ShipmentID, &status, &e.Note, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Status = domain.Status(status)
		events = append(events, &e)
	}
	return events, rows.Err()
}
