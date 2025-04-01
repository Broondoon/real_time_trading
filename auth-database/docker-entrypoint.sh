#!/bin/bash
set -e

# Start PostgreSQL
docker-entrypoint.sh postgres &

# Wait for PostgreSQL to be ready
until PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q'; do
    echo "Postgres is unavailable - sleeping"
    sleep 1
done

echo "Postgres is up - executing cluster setup"
PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f /docker-entrypoint-setupCluster.sql

# Keep container running
wait