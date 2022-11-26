package command

import (
	"cas/backends/s3"
	"cas/localstorage"
	"context"
	"os"
	"testing"

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

func TestStoringArtifacts(t *testing.T) {
	configureTestEnvironment()

	hash := uuid.Must(uuid.NewUUID()).String()

	sourceStore := localstorage.NewMemoryStorage()
	sourceStore.Store["some/path/here"] = []byte("this is the content")
	sourceStore.Store["some/other/path"] = []byte("other content")

	ui := cli.NewMockUi()
	write := &StoreCommand{}
	write.Meta = NewMeta(ui, write)
	write.storage = sourceStore

	// write the file to s3
	assert.Equal(t, 0, write.Run([]string{hash, "some/path/here"}), ui.ErrorWriter.String())
	assert.Equal(t, 0, write.Run([]string{hash, "some/other/path"}), ui.ErrorWriter.String())

	ui.ErrorWriter.Reset()
	ui.OutputWriter.Reset()

	destStore := localstorage.NewMemoryStorage()

	read := &FetchCommand{}
	read.Meta = NewMeta(ui, read)
	read.storage = destStore

	assert.Equal(t, 0, read.Run([]string{hash, "some/path/here"}))

	assert.Equal(t, []byte("this is the content"), destStore.Store["some/path/here"])
	assert.NotContains(t, destStore.Store, "some/other/path")

}
