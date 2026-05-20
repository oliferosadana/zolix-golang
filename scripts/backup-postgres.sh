#!/bin/sh
set -eu

BACKUP_DIR="${BACKUP_DIR:-/data/zolix/backups}"
CONTAINER="${POSTGRES_CONTAINER:-zolix-postgres}"
DB_USER="${POSTGRES_USER:-zolix}"
DB_NAME="${POSTGRES_DB:-zolix}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

mkdir -p "$BACKUP_DIR"

timestamp="$(date +%Y%m%d-%H%M%S)"
target="$BACKUP_DIR/zolix-$timestamp.sql.gz"

docker exec "$CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$target"

find "$BACKUP_DIR" -type f -name 'zolix-*.sql.gz' -mtime +"$RETENTION_DAYS" -delete

echo "$target"
