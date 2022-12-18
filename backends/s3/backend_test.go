package s3

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createConfig() S3Config {

	endpoint := "http://localhost:9000"
	if val := os.Getenv("CAS_S3_TEST_ENDPOINT"); val != "" {
		endpoint = val
	}

	return S3Config{
		Endpoint:   endpoint,
		BucketName: "cas",
		PathPrefix: "tests",
		Region:     "localhost",
		AccessKey:  "minio",
		SecretKey:  "password",
	}
}

func TestWriteMetadata(t *testing.T) {

	hash := uuid.Must(uuid.NewUUID()).String()

	cfg := createConfig()
	EnsureBucket(context.Background(), cfg)

	be := NewS3Backend(cfg)

	assert.NoError(t, be.WriteMetadata(context.Background(), hash, "one", strings.NewReader("something")))

	found, err := be.hasMetadata(context.Background(), hash, "one")
	assert.NoError(t, err)
	assert.True(t, found)
}

func TestWriteMetadataBadBucket(t *testing.T) {
	cfg := createConfig()
	cfg.BucketName = "ewfpweofopwef"

	be := NewS3Backend(cfg)

	err := be.WriteMetadata(context.Background(), "hashone", "one", strings.NewReader("something"))
	assert.ErrorContains(t, err, "NoSuchBucket")
}

func TestListMetadataKeys(t *testing.T) {
	cfg := createConfig()
	EnsureBucket(context.Background(), cfg)

	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg)
	assert.NoError(t, be.WriteMetadata(context.Background(), hash, "one", strings.NewReader("something")))
	assert.NoError(t, be.WriteMetadata(context.Background(), hash, "two", strings.NewReader("something")))

	keys, err := be.listMetadataKeys(context.Background(), hash)
	assert.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "one")
	assert.Contains(t, keys, "two")
}

func TestReadMetadataAll(t *testing.T) {
	cfg := createConfig()
	EnsureBucket(context.Background(), cfg)

	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg)
	assert.NoError(t, be.WriteMetadata(context.Background(), hash, "one", strings.NewReader("something")))
	assert.NoError(t, be.WriteMetadata(context.Background(), hash, "two", strings.NewReader("other thing")))

	meta, err := be.ReadMetadata(context.Background(), hash, []string{})
	assert.NoError(t, err)

	assert.Len(t, meta, 2)
	assert.Equal(t, "something", meta["one"])
	assert.Equal(t, "other thing", meta["two"])
}

func TestReadMetadataSpecific(t *testing.T) {
	cfg := createConfig()
	EnsureBucket(context.Background(), cfg)

	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg)
	assert.NoError(t, be.WriteMetadata(context.Background(), hash, "one", strings.NewReader("something")))
	assert.NoError(t, be.WriteMetadata(context.Background(), hash, "two", strings.NewReader("other thing")))

	meta, err := be.ReadMetadata(context.Background(), hash, []string{"one"})
	assert.NoError(t, err)

	assert.Len(t, meta, 1)
	assert.Equal(t, "something", meta["one"])
}

func TestRelative(t *testing.T) {
	base := "dev/artifact/03dad31909a8617dc00fc1312f3e7fbb076a18c03e5d24ad5a5e43a18f896580"
	key := "dev/artifact/03dad31909a8617dc00fc1312f3e7fbb076a18c03e5d24ad5a5e43a18f896580/flagon"

	name, err := filepath.Rel(base, key)
	assert.NoError(t, err)
	assert.Equal(t, "flagon", name)

}
