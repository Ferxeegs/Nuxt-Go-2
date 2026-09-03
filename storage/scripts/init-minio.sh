#!/bin/sh
set -e

# Must match the underscore-free Docker network alias on template_minio
# (MinIO rejects Host headers that contain "_").
MINIO_HOST="${MINIO_HOST:-template-minio-svc}"
MINIO_ENDPOINT="http://${MINIO_HOST}:9000"

echo "Waiting for MinIO at ${MINIO_ENDPOINT}..."
i=0
until mc alias set local "${MINIO_ENDPOINT}" "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}" >/dev/null 2>&1; do
  i=$((i + 1))
  if [ "$i" -ge 30 ]; then
    echo "MinIO did not become ready in time" >&2
    exit 1
  fi
  sleep 1
done

mc mb --ignore-existing "local/${MINIO_DEFAULT_BUCKET}"
echo "Bucket ${MINIO_DEFAULT_BUCKET} ready."

mc mb --ignore-existing "local/${MINIO_PUBLIC_BUCKET}"
echo "Bucket ${MINIO_PUBLIC_BUCKET} ready."

# Public read for assets that should be served anonymously (logos, etc.).
mc anonymous set download "local/${MINIO_PUBLIC_BUCKET}" || true
echo "MinIO init complete."
