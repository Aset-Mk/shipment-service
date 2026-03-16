.PHONY: generate build test lint docker-up docker-down migrate

PROTO_DIR   := proto/shipment
GEN_DIR     := gen/shipment
BINARY_NAME := shipment-service

# Generate Go code from .proto files.
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc
generate:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/shipment.proto

build:
	go build -o bin/$(BINARY_NAME) ./cmd/server

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	@echo "Running migrations..."
	@for f in internal/infrastructure/postgres/migrations/*.sql; do \
		echo "Applying $$f"; \
		psql "$$DATABASE_URL" -f "$$f"; \
	done
