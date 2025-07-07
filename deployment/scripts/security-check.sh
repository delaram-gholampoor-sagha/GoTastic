#!/bin/bash

if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

echo "Running gosec security checks..."
gosec ./...

echo "Checking for vulnerable dependencies..."
go list -json -m all | nancy sleuth

echo "Checking license compliance..."
go list -m all | go-licenses check

echo "Running static analysis..."
go vet ./... 