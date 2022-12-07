package localstorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListingFiles(t *testing.T) {

	fs := &FileStore{}

	files, err := fs.ListFiles(context.Background(), "testdata")
	assert.NoError(t, err)

	expected := []string{
		"testdata/changelog.md",
		"testdata/collector.yml",
		"testdata/docker-compose.yml",
		"testdata/subdir/test.md",
	}

	assert.Equal(t, expected, files)
}
