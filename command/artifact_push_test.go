package command

import (
	"cas/backends/s3"
	"cas/localstorage"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func configureTestEnvironment() *BackendConfiguration {

	cfg := s3.S3Config{
		BucketName: "cas",
		PathPrefix: "tests",
	}
	s3.EnsureBucket(context.Background(), cfg)

	return &BackendConfiguration{
		name: "s3",
		s3:   cfg,
	}
}

func TestArtifactHashBased(t *testing.T) {
	cfg := configureTestEnvironment()

	now := time.Now()
	hash := uuid.New().String()

	source := localstorage.NewMemoryStorage()
	artifact := NewArtifactPushCommand(localstorage.NewArchiveDecorator(source))
	artifact.backendCfg = cfg

	source.WriteFile(context.Background(), "dist/bin/test", now, strings.NewReader("this is a test"))

	err := artifact.RunContext(context.Background(), []string{hash, "dist/bin/test"})
	assert.NoError(t, err)

	//
	// read back from s3
	//

	dest := localstorage.NewMemoryStorage()
	fetch := NewFetchCommand(localstorage.NewArchiveDecorator(dest))
	fetch.backendCfg = cfg
	fetch.testHash = hash

	err = fetch.RunContext(context.Background(), []string{})
	assert.NoError(t, err)

	assert.Equal(t, []byte("this is a test"), dest.Store["dist/bin/test"])
	assert.WithinRange(t,
		dest.Modified["dist/bin/test"],
		now.Add(-time.Second),
		now.Add(+time.Second))
}
