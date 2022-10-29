# CAS Binary

## Design

Initially designed to work with S3, or a local file system

## Configuration

- Environment variables
- Configuration file?

| Category    | Name            | EnvVar              | Default       | Example                 | Description                                   |
|-------------|-----------------|---------------------|---------------|-------------------------|-----------------------------------------------|
| Common      | Prefix          | `CAS_PREFIX`        | `<empty>`     | `online-web/router`     | A prefix to use in remote state; for segmenting different apps in the same bucket. |
| Common      | Local State     | `CAS_STATE_PATH`    | `.state`      | `./deploy/.state`       | The path to where local copies of state are kept. Used to prevent re-fetching the same artifacts repeatedly. |
| Common      | Remote Backend  | `CAS_BACKEND`       | `s3`          | `fs`                    | The backend to use for remote state storage |
| S3          | Bucket Name     | `CAS_S3_BUCKET`     | `<empty>`     | `eos-artifacts`         | The S3 Bucket to store state in. |
| S3          | Access Key      | `CAS_S3_ACCESS_KEY` | `<empty>`     | `some-access-key`       | S3 Bucket access key (`AWS_ACCESS_KEY`) |
| S3          | Secret Key      | `CAS_S3_SECRET_KEY` | `<empty>`     | `some-access-key`       | S3 Bucket secret key (`AWS_SECRET_ACCESS_KEY`) |
| S3          | Endpoint        | `CAS_S3_ENDPOINT`   | `<empty>`     | `http://localhost:9001` |The S3 endpoint, useful for local testing with Minio. |
| File System | Directory       | `CAS_FS_PATH`       | `/tmp/casfs`  | `../cas`                | A directory to use as a remote state store. |

## CLI

- `artifacts read <hash> [<keyname>,...]`
  - reads all metadata from a hash
  - if `keyname`(s), read only those keys
  - one key + value per line
  - exit is `1` if the `hash` doesn't exist

- `artifacts write <hash> key=value[,key=value...]`
  - writes a key+value to a given `hash`
  - if the `hash` doesn't exist, create it
  - exit is `1` if there is an error

- `artifacts fetch <hash> [<artifact_path>,...]`
  - downloads all artifacts from a hash, to their relative path on disk
  - if a `artifact_path`(s) are given, only download those
  - exit is `1` if the `hash` doesn't exist
  - exit is `2` if a `artifact_path` given doesn't exist
  - `--directory` flag to re-parent all assets
    - given `artifact: some/path/to/file.tar`
    - command: `artifacts fetch <hash> some/path/to/file.tar --directory bin`
    - would write the file to `./bin/file.tar`

- `artifacts store <hash> <artifact_path>[,...]`
  - uploads artifact(s) to storage
  - if the `hash` doesn't exist, create it
