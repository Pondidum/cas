package s3

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteMetadata(t *testing.T) {
	cfg := S3Config{
		Endpoint:   "http://localhost:9000",
		BucketName: "cas",
		PathPrefix: "tests",
		Region:     "localhost",
		AccessKey:  "minio",
		SecretKey:  "password",
	}

	be := NewS3Backend(cfg)

	err := be.WriteMetadata(context.Background(), "hashone", map[string]string{
		"one":       "something",
		"two":       "other thing",
		"timestamp": fmt.Sprintf("%v", (time.Now().Unix())),
	})

	assert.NoError(t, err)
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

	err := be.WriteMetadata(context.Background(), "hashone", map[string]string{
		"one": "something",
		"two": "other thing",
	})

	assert.ErrorContains(t, err, "2 errors occurred:")
}
