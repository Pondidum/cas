# 001 - S3 Storage Layout

- we want to avoid editing files, as this can have race conditions if there is concurrent access
- we can assume full control of our bucket prefix
- we need to store
  - the hash
  - metadata with the hash (in the form of key-value pairs)
  - assets (which have a relative paths)
- we should support being able to delete old hashes, including metadata and artifacts

## Considered Options

### 1.File per metadata key, MRU Cache

```
s3://bucket/prefix
  - meta/
    - {hash}
      - key1  => content is value
      - key2  => content is value
      - ...
      - keyn  => content is value
  - artifacts
    - {hash}
      ./relative/path/is/key => content is value
  - mru
    - {hash}
    - {hash}
```

#### Operations

1. Write k=v for hash:
  ```
  echo "${value}" | aws s3 cp "s3://{bucket}/{prefix}/meta/{hash}/{key}" -
  ```

2. List all keys for hash:
  ```
  aws s3 ls s3://{bucket}/{prefix}/meta/{hash}/
  echo "" | aws s3 cp s3://{bucket}/{prefix}/mru/{hash}
  ```

3. Read value
  ```
  aws s3 cp s3://{bucket}/{prefix}/meta/{hash}/{key} -
  echo "" | aws s3 cp s3://{bucket}/{prefix}/mru/{hash}
  ```

4. store artifact to hash:
  ```
  aws s3 cp ${artifact_relative_path} "s3://{bucket}/{prefix}/artifacts/{hash}/{artifact_relative_path}"
  ```

5. List all artifacts for hash:
  ```
  aws s3 ls s3://{bucket}/{prefix}/artifacts/{hash}/
  echo "" | aws s3 cp s3://{bucket}/{prefix}/mru/{hash}
  ```

6. fetch artifact from hash:
  ```
  aws s3 cp "s3://{bucket}/{prefix}/artifacts/{hash}/{artifact_relative_path}" ${artifact_relative_path}
  echo "" | aws s3 cp s3://{bucket}/{prefix}/mru/{hash}
  ```

7. clean old hashes
  - list hashes in `mru`
  - any older than x time period
  - delete meta/hash/...
  - delete artifacts/hash/...
  - delete mru/hash

#### Considerations

- reading and writing multiple keys is not very fast
- fetching all artifacts should just be an s3 sync command, as we're writing to relative paths
- fetch operations write to the `mru/{hash}` file, so that its modified date gets changed


## Selected Option

[Option 1](#1file-per-metadata-key-mru-cache)

## Considerations

This could be slow depending on usage patterns for fetching all artifacts/metadata.