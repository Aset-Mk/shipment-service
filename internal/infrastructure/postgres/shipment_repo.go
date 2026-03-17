package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Aset-Mk/shipment-service/internal/domain"
)

type shipmentRepo struct {
	db *pgxpool.Pool
}

// NewShipmentRepo creates a Postgres-backed ShipmentRepository.
func NewShipmentRepo(db *pgxpool.Pool) domain.ShipmentRepository {
	return &shipmentRepo{db: db}
}

const saveShipmentSQL = `
INSERT INTO shipments
    (id, reference, origin, destination, status,
     driver_name, driver_license, unit_id, unit_type,
     amount, driver_revenue, created_at, updated_at)
VALUES
    ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

func (r *shipmentRepo) Save(ctx context.Context, s *domain.Shipment) error {
	_, err := r.db.Exec(ctx, saveShipmentSQL,
		s.ID, s.Reference, s.Origin, s.Destination, string(s.Status),
		s.Driver.Name, s.Driver.License,
		s.Unit.ID, s.Unit.Type,
		s.Amount, s.DriverRevenue,
		s.CreatedAt, s.UpdatedAt,
	)
	return err
}

const updateShipmentSQL = `
UPDATE shipments
SET status = $1, updated_at = $2
WHERE id = $3`

func (r *shipmentRepo) Update(ctx context.Context, s *domain.Shipment) error {
	tag, err := r.db.Exec(ctx, updateShipmentSQL, string(s.Status), s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrShipmentNotFound
	}
	return nil
}

const findByIDSQL = `
SELECT id, reference, origin, destination, status,
       driver_name, driver_license, unit_id, unit_type,
       amount, driver_revenue, created_at, updated_at
FROM shipments
WHERE id = $1`

func (r *shipmentRepo) FindByID(ctx context.Context, id string) (*domain.Shipment, error) {
	row := r.db.QueryRow(ctx, findByIDSQL, id)
	return scanShipment(row)
}

const findByReferenceSQL = `
SELECT id, reference, origin, destination, status,
       driver_name, driver_license, unit_id, unit_type,
       amount, driver_revenue, created_at, updated_at
FROM shipments
WHERE reference = $1`

func (r *shipmentRepo) FindByReference(ctx context.Context, ref string) (*domain.Shipment, error) {
	row := r.db.QueryRow(ctx, findByReferenceSQL, ref)
	return scanShipment(row)
}

func scanShipment(row pgx.Row) (*domain.Shipment, error) {
	var s domain.Shipment
	var status string
	err := row.Scan(
		&s.ID, &s.Reference, &s.Origin, &s.Destination, &status,
		&s.Driver.Name, &s.Driver.License,
		&s.Unit.ID, &s.Unit.Type,
		&s.Amount, &s.DriverRevenue,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrShipmentNotFound
		}
		return nil, err
	}
	s.Status = domain.Status(status)
	return &s, nil
}
