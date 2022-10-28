package s3

import (
	"context"
	"testing"

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

	be := NewS3Backend(cfg)

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

	be := NewS3Backend(cfg)

	_, err := be.WriteMetadata(context.Background(), "hashone", map[string]string{
		"one": "something",
		"two": "other thing",
	})

	assert.ErrorContains(t, err, "2 errors occurred:")
}
