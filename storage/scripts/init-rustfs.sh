#!/bin/sh
set -e

# Bucket bootstrap for the S3-compatible object store (RustFS) using aws-cli.
RUSTFS_HOST="${RUSTFS_HOST:-template-rustfs-svc}"
ENDPOINT="http://${RUSTFS_HOST}:9000"

export AWS_ACCESS_KEY_ID="${RUSTFS_ROOT_USER}"
export AWS_SECRET_ACCESS_KEY="${RUSTFS_ROOT_PASSWORD}"
export AWS_DEFAULT_REGION="${RUSTFS_SITE_REGION:-us-east-1}"

s3() { aws --endpoint-url "${ENDPOINT}" "$@"; }

echo "Waiting for object storage at ${ENDPOINT}..."
i=0
until s3 s3api list-buckets >/dev/null 2>&1; do
  i=$((i + 1))
  if [ "$i" -ge 30 ]; then
    echo "Object storage did not become ready in time" >&2
    exit 1
  fi
  sleep 1
done

make_bucket() {
  if s3 s3api head-bucket --bucket "$1" >/dev/null 2>&1; then
    echo "Bucket $1 already exists."
  else
    s3 s3api create-bucket --bucket "$1"
  fi
  echo "Bucket $1 ready."
}

make_bucket "${RUSTFS_DEFAULT_BUCKET}"
make_bucket "${RUSTFS_PUBLIC_BUCKET}"

# Public read for assets that should be served anonymously (logos, etc.).
s3 s3api put-bucket-policy --bucket "${RUSTFS_PUBLIC_BUCKET}" --policy "{
  \"Version\": \"2012-10-17\",
  \"Statement\": [{
    \"Effect\": \"Allow\",
    \"Principal\": {\"AWS\": [\"*\"]},
    \"Action\": [\"s3:GetObject\"],
    \"Resource\": [\"arn:aws:s3:::${RUSTFS_PUBLIC_BUCKET}/*\"]
  }]
}" || echo "Warning: could not set public-read policy" >&2

echo "Object storage init complete."
