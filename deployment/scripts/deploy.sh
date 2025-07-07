#!/bin/bash

set -e

if [ -f .env ]; then
    source .env
fi

echo "Building the application..."
make build

echo "Running security checks..."
./deployment/scripts/security-check.sh

echo "Running tests..."
make test

echo "Building Docker image..."
docker build -t gotastic:latest -f deployment/docker/Dockerfile .

if [ ! -z "$REGISTRY_URL" ]; then
    echo "Pushing to registry..."
    docker tag gotastic:latest $REGISTRY_URL/gotastic:latest
    docker push $REGISTRY_URL/gotastic:latest
fi

echo "Deploying services..."
docker-compose -f deployment/docker/docker-compose.yml up -d

echo "Deployment completed successfully!" 