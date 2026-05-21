#!/bin/sh
set -eu

BACKUP_DIR="${BACKUP_DIR:-/data/zolix/backups}"
DB_HOST="${POSTGRES_HOST:-127.0.0.1}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-zolix_user}"
DB_NAME="${POSTGRES_DB:-zolix_db}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"

mkdir -p "$BACKUP_DIR"

timestamp="$(date +%Y%m%d-%H%M%S)"
target="$BACKUP_DIR/zolix-$timestamp.sql.gz"

pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" | gzip > "$target"

find "$BACKUP_DIR" -type f -name 'zolix-*.sql.gz' -mtime +"$RETENTION_DAYS" -delete

echo "$target"
