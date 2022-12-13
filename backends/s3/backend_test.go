package s3

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

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
	cfg := createConfig()
	cfg.BucketName = "ewfpweofopwef"

	be := NewS3Backend(cfg)

	_, err := be.WriteMetadata(context.Background(), "hashone", map[string]string{
		"one": "something",
		"two": "other thing",
	})

	assert.ErrorContains(t, err, "3 errors occurred:")
}

func TestListMetadataKeys(t *testing.T) {
	cfg := createConfig()
	EnsureBucket(context.Background(), cfg)

	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg)
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
	cfg := createConfig()
	EnsureBucket(context.Background(), cfg)

	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg)
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
	cfg := createConfig()
	EnsureBucket(context.Background(), cfg)

	hash := uuid.Must(uuid.NewUUID()).String()

	be := NewS3Backend(cfg)
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

func TestRelative(t *testing.T) {
	base := "dev/artifact/03dad31909a8617dc00fc1312f3e7fbb076a18c03e5d24ad5a5e43a18f896580"
	key := "dev/artifact/03dad31909a8617dc00fc1312f3e7fbb076a18c03e5d24ad5a5e43a18f896580/flagon"

	name, err := filepath.Rel(base, key)
	assert.NoError(t, err)
	assert.Equal(t, "flagon", name)

}
