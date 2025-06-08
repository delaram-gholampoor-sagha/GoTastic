SHELL := /bin/bash

.PHONY: build run test clean migrate-up migrate-down init-s3 benchmark deploy dev security-check

# Default target
all: build

# Build the application
build:
	go build -o bin/api cmd/api/main.go

# Run the application
run:
	docker-compose -f deployment/docker/docker-compose.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services are ready!"

# Run in development mode
dev:
	docker-compose -f deployment/docker/docker-compose.dev.yml up -d
	@echo "Development environment is ready!"

# Run tests
test:
	go test -v -timeout 30s ./... -run "^Test"

# Run benchmarks
benchmark:
	go test -bench=. -benchmem -timeout 30s ./...

# Clean build artifacts
clean:
	docker-compose -f deployment/docker/docker-compose.yml down -v
	rm -rf bin/
	rm -rf tmp/

# Run database migrations up
migrate-up:
	migrate -path migrations -database "mysql://root:password@tcp(localhost:3306)/todo" up

# Run database migrations down
migrate-down:
	migrate -path migrations -database "mysql://root:password@tcp(localhost:3306)/todo" down

# Initialize S3 bucket
init-s3:
	aws --endpoint-url http://localhost:4570 s3 mb s3://todo-files

# Run security checks
security-check:
	./deployment/scripts/security-check.sh

# Deploy the application
deploy:
	./deployment/scripts/deploy.sh

# Help
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make dev           - Run in development mode with hot-reload"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean up"
	@echo "  make migrate-up    - Run database migrations up"
	@echo "  make migrate-down  - Run database migrations down"
	@echo "  make init-s3       - Initialize S3 bucket"
	@echo "  make benchmark     - Run benchmarks"
	@echo "  make security-check - Run security checks"
	@echo "  make deploy        - Deploy the application" 