#!/bin/bash
set -e

CONTAINER_NAME="pg_schema_diff_demo"
DB_NAME="testdb"
DB_USER="postgres"
DB_PASS="postgres"

cleanup() {
    echo "Cleaning up..."
    docker rm -f $CONTAINER_NAME 2>/dev/null || true
}

trap cleanup EXIT
cleanup

echo "=== pg-schema-diff Expression Index Failure Demo ==="
echo "Starting Postgres container..."
docker run -d --name $CONTAINER_NAME \
    -e POSTGRES_USER=$DB_USER \
    -e POSTGRES_PASSWORD=$DB_PASS \
    -e POSTGRES_DB=$DB_NAME \
    -p 5433:5432 \
    postgres:15 > /dev/null

echo "Waiting for Postgres to be ready..."
until docker exec $CONTAINER_NAME pg_isready -U $DB_USER -d $DB_NAME > /dev/null 2>&1; do
    sleep 1
done
sleep 2

echo "Applying schema_no_index.sql (table only)..."
docker exec -i $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME < schema_no_index.sql

export DATABASE_URL="postgresql://${DB_USER}:${DB_PASS}@localhost:5433/${DB_NAME}"
go run .
