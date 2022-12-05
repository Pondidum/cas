package localstorage

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStoreListingFiles(t *testing.T) {
	ctx := context.Background()

	store := NewMemoryStorage()

	store.WriteFile(ctx, "root/child/one.js", strings.NewReader("one"))
	store.WriteFile(ctx, "root/child/two.js", strings.NewReader("two"))
	store.WriteFile(ctx, "root/child/three.js", strings.NewReader("three"))
	store.WriteFile(ctx, "root/child/four.js", strings.NewReader("four"))
	store.WriteFile(ctx, "root/a.js", strings.NewReader("a"))
	store.WriteFile(ctx, "root/b.js", strings.NewReader("b"))

	files, err := store.ListFiles(ctx, "root")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"root/a.js",
		"root/b.js",
		"root/child/four.js",
		"root/child/one.js",
		"root/child/three.js",
		"root/child/two.js",
	}, files)
}

func TestMemoryStorageReadWriteFiles(t *testing.T) {
	ctx := context.Background()

	store := NewMemoryStorage()

	store.WriteFile(ctx, "root/child/one.js", strings.NewReader("one"))
	store.WriteFile(ctx, "root/child/two.js", strings.NewReader("two"))

	bytes, err := store.ReadFile(ctx, "root/child/one.js")
	assert.NoError(t, err)

	content, _ := ioutil.ReadAll(bytes)
	assert.Equal(t, []byte("one"), content)
}
