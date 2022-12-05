package localstorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListingFiles(t *testing.T) {

	fs := &FileStore{}

	files, err := fs.ListFiles(context.Background(), ".")
	assert.NoError(t, err)

	expected := []string{

		"archive.go",
		"directory_handler.go",
		"directory_handler_test.go",
		"filestore.go",
		"filestore_test.go",
		"memorystore.go",
		"storage.go",
		"testdata/changelog.md",
		"testdata/collector.yml",
		"testdata/docker-compose.yml",
		"testdata/subdir/test.md",
	}

	assert.Equal(t, expected, files)
}
