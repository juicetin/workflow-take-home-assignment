#!/bin/bash

# Migration script for CI/CD environments
# Usage: ./scripts/migrate.sh [up|version] [database-url]

set -e

COMMAND=${1:-up}
DATABASE_URL=${2:-$DATABASE_URL}

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL must be provided as environment variable or second argument"
    echo "Usage: $0 [up|version] [database-url]"
    exit 1
fi

echo "Running database migrations..."
echo "Command: $COMMAND"
echo "Database URL: ${DATABASE_URL%@*}@***" # Hide credentials in logs

# Build and run the migration tool
cd "$(dirname "$0")/.."
go run ./scripts/migrate.go -command="$COMMAND" -database-url="$DATABASE_URL" -migrations-path="./migrations"

echo "Migration completed successfully"