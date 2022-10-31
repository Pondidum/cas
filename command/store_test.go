package command

import (
	"bytes"
	"cas/backends/s3"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

type closableBuffer struct {
	*bytes.Reader
}

func (b *closableBuffer) Close() error {
	return nil
}

func createFs(fs map[string][]byte) (s3.ReadFunc, s3.WriteFunc) {

	r := func(ctx context.Context, p string) (io.ReadCloser, error) {
		if content, found := fs[p]; found {
			return &closableBuffer{Reader: bytes.NewReader(content)}, nil
		}

		return nil, fmt.Errorf("file not found: %s", p)
	}

	w := func(ctx context.Context, path string, content io.Reader) (string, error) {

		b, err := ioutil.ReadAll(content)
		if err != nil {
			return "", err
		}

		fs[path] = b

		return path, nil
	}

	return r, w
}

func TestStoringArtifacts(t *testing.T) {

	os.Setenv("CAS_S3_ENDPOINT", "http://localhost:9000")
	os.Setenv("CAS_S3_REGION", "localhost")
	os.Setenv("CAS_S3_BUCKET", "cas")
	os.Setenv("CAS_S3_ACCESS_KEY", "minio")
	os.Setenv("CAS_S3_SECRET_KEY", "password")
	os.Setenv("CAS_S3_PATH_PREFIX", "cli")

	hash := uuid.Must(uuid.NewUUID()).String()

	sourceReader, _ := createFs(map[string][]byte{
		"some/path/here":  []byte("this is the content"),
		"some/other/path": []byte("other content"),
	})

	ui := cli.NewMockUi()
	write := &StoreCommand{}
	write.Meta = NewMeta(ui, write)
	write.Meta.customRead = sourceReader

	// write the file to s3
	assert.Equal(t, 0, write.Run([]string{hash, "some/path/here"}), ui.ErrorWriter.String())
	assert.Equal(t, 0, write.Run([]string{hash, "some/other/path"}), ui.ErrorWriter.String())

	ui.ErrorWriter.Reset()
	ui.OutputWriter.Reset()

	destFs := map[string][]byte{}
	_, destWriter := createFs(destFs)

	read := &FetchCommand{}
	read.Meta = NewMeta(ui, read)
	read.Meta.customWrite = destWriter

	assert.Equal(t, 0, read.Run([]string{hash, "some/path/here"}))

	assert.Equal(t, []byte("this is the content"), destFs["some/path/here"])
	assert.NotContains(t, destFs, "some/other/path")

}
