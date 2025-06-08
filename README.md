# GoTastic - Todo Service

A clean architecture implementation of a Todo service with file upload capabilities, using MySQL, Redis Streams, and S3.

## Prerequisites

- Docker
- Docker Compose
- Go 1.23 or newer
- Make
- AWS CLI (for S3 operations)

## Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/delaram/GoTastic.git
   cd GoTastic
   ```

2. **Start the application using Docker Compose (recommended):**
   ```bash
   make dev
   ```
   This will start MySQL, Redis, LocalStack (for S3), and the API server.

3. **Run database migrations:**
   ```bash
   make migrate-up
   ```

4. **Create the S3 bucket (if not already created):**
   ```bash
   docker exec -it -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test localstack aws --endpoint-url=http://localhost:4566 s3 mb s3://todo-files
   ```

5. **(Optional) Run the API locally:**
   If you want to run the API outside Docker, set these environment variables:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=3306
   export DB_USER=root
   export DB_PASSWORD=password
   export DB_NAME=todo
   go run cmd/api/main.go
   ```

## Testing the API

### Health Check

```bash
curl -i http://localhost:8080/health
```

### File Upload

```bash
curl -F "file=@test.txt" http://localhost:8080/api/v1/files/
```

### Todo Creation

```bash
curl -X POST -H "Content-Type: application/json" -d '{"description":"Test todo","due_date":"2025-12-31T00:00:00Z"}' http://localhost:8080/api/v1/todos/
```

### List Todos

```bash
curl http://localhost:8080/api/v1/todos/
```

### Get Todo by ID

```bash
curl http://localhost:8080/api/v1/todos/<todo-id>
```

### Update Todo

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"description":"Updated description","due_date":"2025-12-31T00:00:00Z"}' http://localhost:8080/api/v1/todos/<todo-id>
```

### Delete Todo

```bash
curl -X DELETE http://localhost:8080/api/v1/todos/<todo-id>
```

### Download File

```bash
curl http://localhost:8080/api/v1/files/<file-id>
```

### Delete File

```bash
curl -X DELETE http://localhost:8080/api/v1/files/<file-id>
```

## Troubleshooting

- **API returns 404 or 500:** Ensure all containers are running and migrations have been applied.
- **File upload fails:** Make sure the S3 bucket exists in LocalStack.
- **Database errors:** Confirm the correct environment variables and that the `todo_items` table exists.

## Development

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/usecase/...

# Run tests with coverage
go test -cover ./...
```

### Running Benchmarks

```bash
# Run all benchmarks
make benchmark

# Run specific package benchmarks
go test -bench=. -benchmem ./internal/usecase/...
```

### Development Mode

```bash
# Start services with hot-reload
make dev
```

## Project Structure

```
├── cmd/                    # Application entry points
├── config/                 # Configuration files
├── deployment/            # Deployment configurations
│   ├── docker/           # Docker files
│   └── scripts/          # Deployment scripts
├── internal/             # Private application code
│   ├── config/          # Configuration package
│   ├── delivery/        # HTTP handlers
│   ├── domain/          # Domain models
│   ├── repository/      # Data access layer
│   └── usecase/         # Business logic
├── migrations/           # Database migrations
├── pkg/                 # Shared packages
└── Makefile            # Build commands
```

## Clean Architecture

The project follows Clean Architecture principles:
- Domain Layer: Core business logic and entities
- Use Case Layer: Application-specific business rules
- Interface Adapters: Controllers, Gateways, Presenters
- Frameworks & Drivers: External frameworks and tools

## Testing Strategy

- Unit Tests: Test individual components in isolation
- Integration Tests: Test component interactions
- Benchmarks: Measure performance of critical operations

### Mocking Strategy

- Repository interfaces for database operations
- S3 client interface for file operations
- Redis client interface for stream operations

## Performance Considerations

- Redis caching for frequently accessed data
- Connection pooling for database and Redis
- Efficient file handling with streaming

## License

This project is for demonstration purposes only.