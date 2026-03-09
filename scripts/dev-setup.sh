#!/bin/bash
set -e

# Always run from the repo root regardless of where the script is called from
cd "$(dirname "$0")/.."

echo "Starting core_db..."
podman compose up core_db -d

echo "Waiting for database to be ready..."
until podman compose exec core_db pg_isready -U core -d secure_registry -q; do
  sleep 1
done
echo "Database is ready."

echo "Running migrations..."

podman compose exec -T core_db psql -U core -d secure_registry -q \
  < infra/migrations/000001_initial_schema.up.sql

podman compose exec -T core_db psql -U core -d secure_registry -q \
  < infra/migrations/000002_better_auth_organization.up.sql

podman compose exec -T core_db psql -U core -d secure_registry -q \
  < infra/migrations/000003_add_auth_tables.up.sql

echo "Migrations complete."
echo ""
echo "Run 'cd home-ui && bun run dev' to start the app."
