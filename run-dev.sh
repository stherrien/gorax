#!/bin/bash

# Load .env file
set -a
source .env
set +a

# Run API in development mode
echo "Starting API in development mode..."
echo "Database: ${DB_NAME} at ${DB_HOST}:${DB_PORT}"
go run ./cmd/api
