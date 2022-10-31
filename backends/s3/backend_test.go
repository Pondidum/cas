package s3

import (
	"cas/backends"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWriteMetadata(t *testing.T) {

	hash := uuid.Must(uuid.NewUUID()).String()

	cfg := S3Config{
		Endpoint:   "http://localhost:9000",
		BucketName: "cas",
		PathPrefix: "tests",
		Region:     "localhost",
		AccessKey:  "minio",
		SecretKey:  "password",
	}

	be := NewS3Backend(cfg, backends.NewMemoryStorage())

	written, err := be.WriteMetadata(context.Background(), hash, map[string]string{
		"one": "something",
		"two": "other thing",
	})

	assert.NoError(t, err)
	assert.Contains(t, written, "one")
	assert.Contains(t, written, "two")
	assert.Contains(t, written, MetadataTimeStamp)

	found, err := be.hasMetadata(context.Background(), hash, "@timestamp")
	assert.NoError(t, err)
	assert.True(t, found)
}

func TestWriteMetadataBadBucket(t *testing.T) {
	cfg := S3Config{
		Endpoint:   "http://localhost:9000",
		BucketName: "efjwoeoijweoifj",
		PathPrefix: "tests",
		Region:     "localhost",
		AccessKey:  "minio",
		SecretKey:  "password",
	}

	be := NewS3Backend(cfg, backends.NewMemoryStorage())

	_, err := be.WriteMetadata(context.Background(), "hashone", map[string]string{
		"one": "something",
		"two": "other thing",
	})

	assert.ErrorContains(t, err, "3 errors occurred:")
}

func TestListMetadataKeys(t *testing.T) {
	cfg := S3Config{
		Endpoint:   "http://localhost:9000",
		BucketName: "cas",
		PathPrefix: "tests",
		Region:     "localhost",
		AccessKey:  "minio",
		SecretKey:  "password",
	}
	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg, backends.NewMemoryStorage())
	_, err := be.WriteMetadata(context.Background(), hash, map[string]string{
		"one": "something",
		"two": "other thing",
	})
	assert.NoError(t, err)

	keys, err := be.listMetadataKeys(context.Background(), hash)
	assert.NoError(t, err)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "one")
	assert.Contains(t, keys, "two")
	assert.Contains(t, keys, MetadataTimeStamp)
}

func TestReadMetadataAll(t *testing.T) {
	cfg := S3Config{
		Endpoint:   "http://localhost:9000",
		BucketName: "cas",
		PathPrefix: "tests",
		Region:     "localhost",
		AccessKey:  "minio",
		SecretKey:  "password",
	}
	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg, backends.NewMemoryStorage())
	_, err := be.WriteMetadata(context.Background(), hash, map[string]string{
		"one": "something",
		"two": "other thing",
	})
	assert.NoError(t, err)

	meta, err := be.ReadMetadata(context.Background(), hash, []string{})
	assert.NoError(t, err)

	assert.Len(t, meta, 3)
	i, _ := strconv.Atoi(meta[MetadataTimeStamp])
	assert.Equal(t, "something", meta["one"])
	assert.Equal(t, "other thing", meta["two"])
	assert.InDelta(t, time.Now().Unix(), i, 10)
}

func TestReadMetadataSpecific(t *testing.T) {
	cfg := S3Config{
		Endpoint:   "http://localhost:9000",
		BucketName: "cas",
		PathPrefix: "tests",
		Region:     "localhost",
		AccessKey:  "minio",
		SecretKey:  "password",
	}
	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg, backends.NewMemoryStorage())
	_, err := be.WriteMetadata(context.Background(), hash, map[string]string{
		"one": "something",
		"two": "other thing",
	})
	assert.NoError(t, err)

	meta, err := be.ReadMetadata(context.Background(), hash, []string{"one", MetadataTimeStamp})
	assert.NoError(t, err)

	assert.Len(t, meta, 2)
	i, _ := strconv.Atoi(meta[MetadataTimeStamp])
	assert.Equal(t, "something", meta["one"])
	assert.InDelta(t, time.Now().Unix(), i, 10)
}
