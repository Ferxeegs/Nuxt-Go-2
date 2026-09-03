#!/bin/sh
# Blocks until the given MySQL host:port accepts TCP connections. Useful as
# an entrypoint prefix when the shared MySQL container starts slower than
# this app container (the Go app itself also retries internally, see
# internal/database/connection.go, but this gives a faster fail-fast option
# for scripted/CI use).
set -e

HOST="${1:-$DB_HOST}"
PORT="${2:-$DB_PORT}"
TIMEOUT="${3:-60}"

echo "waiting for mysql at ${HOST}:${PORT} (timeout ${TIMEOUT}s)..."

elapsed=0
until nc -z "$HOST" "$PORT" >/dev/null 2>&1; do
  elapsed=$((elapsed + 1))
  if [ "$elapsed" -ge "$TIMEOUT" ]; then
    echo "timed out waiting for mysql at ${HOST}:${PORT}"
    exit 1
  fi
  sleep 1
done

echo "mysql is up at ${HOST}:${PORT}"
