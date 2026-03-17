# Shipment Service

A gRPC microservice for managing shipments and tracking their status changes.

## Running the service

### With Docker (recommended)

```bash
docker-compose up --build
```

The service starts on `:50051`. Postgres migrations run automatically on first launch.

### Without Docker

1. Start a Postgres instance and apply migrations manually:

```bash
psql $DATABASE_URL -f internal/infrastructure/postgres/migrations/001_create_shipments.sql
psql $DATABASE_URL -f internal/infrastructure/postgres/migrations/002_create_shipment_events.sql
```

2. Set environment variables (copy from `.env.example`):

```bash
export DATABASE_URL=postgres://shipment:shipment@localhost:5432/shipment?sslmode=disable
export GRPC_ADDR=:50051
```

3. Build and run:

```bash
make build
./bin/shipment-service
```

## Running the tests

```bash
make test
```

Or directly:

```bash
go test ./...
```

Tests cover domain logic and use-case behavior. They run without any external dependencies — no database, no network.

## Calling the service

The server has gRPC reflection enabled, so you can use [grpcurl](https://github.com/fullstorydev/grpcurl):

```bash
# create a shipment
grpcurl -plaintext -d '{
  "reference": "REF-001",
  "origin": "Almaty",
  "destination": "Astana",
  "driver": {"name": "Ali Bekov", "license": "AA1234"},
  "unit": {"id": "TRUCK-01", "type": "truck"},
  "amount": 1500.00,
  "driver_revenue": 300.00
}' localhost:50051 shipment.ShipmentService/CreateShipment

# add a status event
grpcurl -plaintext -d '{
  "shipment_id": "<id>",
  "status": "SHIPMENT_STATUS_PICKED_UP",
  "note": "driver arrived at warehouse"
}' localhost:50051 shipment.ShipmentService/AddEvent

# get event history
grpcurl -plaintext -d '{"shipment_id": "<id>"}' \
  localhost:50051 shipment.ShipmentService/GetEvents
```

## Architecture

The project follows Clean Architecture principles. Dependencies flow strictly inward:

```
transport/grpc  →  usecase  →  domain
infrastructure  →  domain
```

```
.
├── cmd/server                        # entry point, dependency wiring
├── internal/
│   ├── domain/                       # business logic, no external deps
│   │   ├── shipment.go               # shipment aggregate
│   │   ├── status.go                 # status type + transition rules
│   │   ├── event.go                  # status event record
│   │   ├── errors.go                 # domain errors
│   │   └── repository.go             # repository interfaces
│   ├── usecase/                      # application logic
│   │   ├── shipment.go               # use-case interface
│   │   └── shipment_service.go       # orchestration
│   ├── infrastructure/postgres/      # repository implementations
│   ├── transport/grpc/               # gRPC server, handlers, proto↔domain mappers
│   └── config/                       # env-based configuration
├── proto/shipment/                   # .proto definitions
└── gen/shipment/                     # generated protobuf code
```

**Domain layer** (`internal/domain`) has zero dependencies outside the standard library. It owns the `Shipment` aggregate, the status FSM, and the repository interfaces. Everything else depends on it — not the other way around.

**Use-case layer** (`internal/usecase`) coordinates domain objects and repositories. It does not know about gRPC, HTTP, or Postgres. Tests here use in-memory repository implementations — no database needed.

**Infrastructure layer** (`internal/infrastructure/postgres`) implements the repository interfaces from the domain using `pgx/v5`. Swapping the database means replacing only this package.

**Transport layer** (`internal/transport/grpc`) handles serialization and maps gRPC requests to use-case calls. Errors from the domain are translated to appropriate gRPC status codes.

## Design decisions

**Status machine defined in the domain.** The `allowedTransitions` map lives in `status.go` and is the single place that controls what status changes are legal. Adding a new status means updating that map and nothing else.

**`ApplyEvent` on the aggregate.** The shipment enforces its own invariants — the use-case does not check transition validity itself, it just calls `shipment.ApplyEvent` and propagates the error. This keeps the business rule close to the data it protects.

**IDGenerator injected as a function.** `NewShipmentService` accepts a `func() string` for generating IDs. This makes the use-case testable with predictable IDs without mocking a package-level function.

**`now` also injected.** The service holds a `func() time.Time` field (defaults to `time.Now`). Tests can override it if time-sensitive assertions are needed.

**Migrations mounted into `docker-entrypoint-initdb.d`.** Postgres runs them automatically on first boot. For a real project I would use a dedicated migration tool (e.g. goose or migrate), but for this scope mounting SQL files is simpler and has fewer moving parts.

**gRPC reflection enabled.** Makes the service explorable with `grpcurl` without distributing the `.proto` files separately.

## Assumptions

- A shipment reference is unique across the system. Attempting to create two shipments with the same reference returns `AlreadyExists`.
- Status values are stored as plain strings in Postgres. This keeps migrations simple and avoids a dependency on database enums when the status list changes.
- `driver_revenue` and `amount` are stored as `NUMERIC(12,2)` — enough precision for logistics costs without floating-point rounding issues.
- The event history is append-only. There is no API to delete or edit past events.
- Authentication and authorization are out of scope for this task.
