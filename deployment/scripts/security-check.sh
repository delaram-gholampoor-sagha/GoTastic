#!/bin/bash

# Install required tools if not present
if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Run security checks
echo "Running gosec security checks..."
gosec ./...

# Run dependency vulnerability check
echo "Checking for vulnerable dependencies..."
go list -json -m all | nancy sleuth

# Run license compliance check
echo "Checking license compliance..."
go list -m all | go-licenses check

# Run static analysis
echo "Running static analysis..."
go vet ./... 