#!/bin/bash

set -e

# Load environment variables
if [ -f .env ]; then
    source .env
fi

# Build the application
echo "Building the application..."
make build

# Run security checks
echo "Running security checks..."
./deployment/scripts/security-check.sh

# Run tests
echo "Running tests..."
make test

# Build Docker image
echo "Building Docker image..."
docker build -t gotastic:latest -f deployment/docker/Dockerfile .

# Push to registry if REGISTRY_URL is set
if [ ! -z "$REGISTRY_URL" ]; then
    echo "Pushing to registry..."
    docker tag gotastic:latest $REGISTRY_URL/gotastic:latest
    docker push $REGISTRY_URL/gotastic:latest
fi

# Deploy using docker-compose
echo "Deploying services..."
docker-compose -f deployment/docker/docker-compose.yml up -d

echo "Deployment completed successfully!" 