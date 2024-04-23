Setup

```bash
set -x CAS_S3_ENDPOINT http://localhost:9000
set -x CAS_S3_BUCKET makestate
set -x CAS_S3_ACCESS_KEY minio
set -x CAS_S3_SECRET_KEY password

AWS_ACCESS_KEY_ID="${CAS_S3_ACCESS_KEY}" AWS_SECRET_ACCESS_KEY="${CAS_S3_SECRET_KEY}" aws --endpoint-url "${CAS_S3_ENDPOINT}" s3 mb "s3://${CAS_S3_BUCKET}"
```