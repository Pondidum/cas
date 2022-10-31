package command

import (
	"cas/backends"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

func TestStoringArtifacts(t *testing.T) {

	os.Setenv("CAS_S3_ENDPOINT", "http://localhost:9000")
	os.Setenv("CAS_S3_REGION", "localhost")
	os.Setenv("CAS_S3_BUCKET", "cas")
	os.Setenv("CAS_S3_ACCESS_KEY", "minio")
	os.Setenv("CAS_S3_SECRET_KEY", "password")
	os.Setenv("CAS_S3_PATH_PREFIX", "cli")

	hash := uuid.Must(uuid.NewUUID()).String()

	sourceStore := backends.NewMemoryStorage()
	sourceStore.Store["some/path/here"] = []byte("this is the content")
	sourceStore.Store["some/other/path"] = []byte("other content")

	ui := cli.NewMockUi()
	write := &StoreCommand{}
	write.Meta = NewMeta(ui, write)
	write.Meta.customStorage = sourceStore

	// write the file to s3
	assert.Equal(t, 0, write.Run([]string{hash, "some/path/here"}), ui.ErrorWriter.String())
	assert.Equal(t, 0, write.Run([]string{hash, "some/other/path"}), ui.ErrorWriter.String())

	ui.ErrorWriter.Reset()
	ui.OutputWriter.Reset()

	destStore := backends.NewMemoryStorage()

	read := &FetchCommand{}
	read.Meta = NewMeta(ui, read)
	read.Meta.customStorage = destStore

	assert.Equal(t, 0, read.Run([]string{hash, "some/path/here"}))

	assert.Equal(t, []byte("this is the content"), destStore.Store["some/path/here"])
	assert.NotContains(t, destStore.Store, "some/other/path")

}
