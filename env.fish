
set -x CAS_S3_PATH_PREFIX dev
set -x CAS_S3_BUCKET makestate
set -x CAS_S3_ENDPOINT "http://localhost:9000"
set -x CAS_S3_REGION localhost
set -x CAS_S3_ACCESS_KEY minio
set -x CAS_S3_SECRET_KEY password

set -x OTEL_TRACE_EXPORTER "otlp"

set -x AWS_SECRET_ACCESS_KEY password
set -x AWS_ACCESS_KEY_ID minio