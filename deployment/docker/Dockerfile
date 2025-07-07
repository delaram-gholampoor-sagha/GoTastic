# Build stage
FROM golang:latest AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y git make bash

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bin/api cmd/api/main.go

FROM debian:stable-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/api /app/api

COPY --from=builder /app/config /app/config

ENV GIN_MODE=release
ENV PORT=8080

EXPOSE 8080

CMD ["./api"] 