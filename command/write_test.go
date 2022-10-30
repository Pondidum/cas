package command

import (
	"cas/backends/s3"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

func TestWriting(t *testing.T) {

	write := &WriteCommand{}
	write.Meta = NewMeta(cli.NewMockUi(), write)

	os.Setenv("CAS_S3_ENDPOINT", "http://localhost:9000")
	os.Setenv("CAS_S3_REGION", "localhost")
	os.Setenv("CAS_S3_BUCKET", "cas")
	os.Setenv("CAS_S3_ACCESS_KEY", "minio")
	os.Setenv("CAS_S3_SECRET_KEY", "password")
	os.Setenv("CAS_S3_PATH_PREFIX", "cli")

	hash := uuid.Must(uuid.NewUUID()).String()

	assert.Equal(t, 0, write.Run([]string{hash, "name=andy", "employer=reaktor"}))

	ui := cli.NewMockUi()
	read := &ReadCommand{}
	read.Meta = NewMeta(ui, read)

	// verify all keys are returned
	assert.Equal(t, 0, read.Run([]string{hash}))
	output := ui.OutputWriter.String()
	assert.Contains(t, output, "name: andy")
	assert.Contains(t, output, "employer: reaktor")
	assert.Contains(t, output, s3.MetadataTimeStamp+": ")

	// verify only specified keys are returned
	ui.OutputWriter.Reset()
	ui.ErrorWriter.Reset()
	assert.Equal(t, 0, read.Run([]string{hash, "employer"}))
	output = ui.OutputWriter.String()
	assert.NotContains(t, output, "name: andy")
	assert.Contains(t, output, "employer: reaktor")
	assert.NotContains(t, output, s3.MetadataTimeStamp+": ")
}
