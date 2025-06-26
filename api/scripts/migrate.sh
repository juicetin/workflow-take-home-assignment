#!/bin/bash

# Migration script for CI/CD environments
# Usage: ./scripts/migrate.sh [up|version] [database-url] [--seed]

set -e

COMMAND=${1:-up}
DATABASE_URL=${2:-$DATABASE_URL}
SEED_FLAG=""

# Check if --seed flag is provided
if [ "$3" == "--seed" ]; then
    SEED_FLAG="-seed"
fi

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL must be provided as environment variable or second argument"
    echo "Usage: $0 [up|version] [database-url] [--seed]"
    exit 1
fi

echo "Running database migrations..."
echo "Command: $COMMAND"
echo "Database URL: ${DATABASE_URL%@*}@***" # Hide credentials in logs
if [ -n "$SEED_FLAG" ]; then
    echo "Seeding: enabled"
fi

# Build and run the migration tool
cd "$(dirname "$0")/.."
go run ./scripts/migrate.go ./scripts/seed_test_data.go -command="$COMMAND" -database-url="$DATABASE_URL" -migrations-path="./migrations" $SEED_FLAG

echo "Migration completed successfully"