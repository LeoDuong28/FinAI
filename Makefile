.PHONY: dev air templ-watch tailwind-watch sqlc templ migrate-up migrate-down migrate-create lint test test-cover build up down clean

# Start everything with hot reload
dev:
	docker compose up -d postgres
	@echo "Waiting for postgres..."
	@sleep 2
	$(MAKE) -j3 air templ-watch tailwind-watch

# Go hot reload
air:
	air

# Templ file watcher
templ-watch:
	templ generate --watch

# Tailwind CSS watcher
tailwind-watch:
	npx tailwindcss -i static/css/app.css -o static/css/output.css --watch

# Generate type-safe DB code from SQL
sqlc:
	sqlc generate

# Generate Templ files
templ:
	templ generate

# Run all pending migrations
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

# Rollback last migration
migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

# Create new migration: make migrate-create NAME=add_feature
migrate-create:
	migrate create -ext sql -dir migrations -seq $(NAME)

# Lint Go code
lint:
	golangci-lint run ./...

# Run all tests
test:
	go test ./... -v -race

# Run tests with coverage
test-cover:
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Build production binary
build:
	templ generate
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/server ./cmd/server

# Docker compose up (all services)
up:
	docker compose up -d

# Docker compose down
down:
	docker compose down

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html
	rm -rf internal/repository/postgres/generated/
