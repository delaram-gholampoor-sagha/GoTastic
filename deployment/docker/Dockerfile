# Build stage
FROM golang:latest AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y git make bash

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o bin/api cmd/api/main.go

# Final stage
FROM debian:stable-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/bin/api /app/api

# Copy configuration
COPY --from=builder /app/config /app/config

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"] 