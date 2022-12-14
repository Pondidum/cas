package command

import (
	"cas/backends/s3"
	"cas/localstorage"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

func configureTestEnvironment() {

	endpoint := "http://localhost:9000"
	if val := os.Getenv("CAS_S3_TEST_ENDPOINT"); val != "" {
		endpoint = val
	}

	os.Setenv("CAS_S3_ENDPOINT", endpoint)
	os.Setenv("CAS_S3_REGION", "localhost")
	os.Setenv("CAS_S3_BUCKET", "cas")
	os.Setenv("CAS_S3_ACCESS_KEY", "minio")
	os.Setenv("CAS_S3_SECRET_KEY", "password")
	os.Setenv("CAS_S3_PATH_PREFIX", "tests")

	cfg := s3.ConfigFromEnvironment()
	s3.EnsureBucket(context.Background(), cfg)
}

func TestArtifactHashBased(t *testing.T) {
	configureTestEnvironment()

	now := time.Now()
	hash := uuid.New().String()
	ui := cli.NewMockUi()

	source := localstorage.NewMemoryStorage()
	artifact := NewArtifactCommand(ui)
	artifact.storage = source
	artifact.backendName = "s3"

	source.WriteFile(context.Background(), "dist/bin/test", now, strings.NewReader("this is a test"))

	err := artifact.RunContext(context.Background(), []string{hash, "dist/bin/test"})
	assert.NoError(t, err)

	//
	// read back from s3
	//

	dest := localstorage.NewMemoryStorage()
	fetch := NewFetchCommand(ui)
	fetch.storage = dest
	fetch.backendName = "s3"
	fetch.testHash = hash

	err = fetch.RunContext(context.Background(), []string{})
	assert.NoError(t, err)

	assert.Equal(t, []byte("this is a test"), dest.Store["dist/bin/test"])
	assert.WithinRange(t,
		dest.Modified["dist/bin/test"],
		now.Add(-time.Second),
		now.Add(+time.Second))
}
