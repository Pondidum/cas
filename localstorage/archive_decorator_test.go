package localstorage

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchiving(t *testing.T) {
	ctx := context.Background()

	source := NewMemoryStorage()
	source.WriteFile(ctx, "test/one.md", strings.NewReader("first file"))
	source.WriteFile(ctx, "test/two.md", strings.NewReader("second file"))
	source.WriteFile(ctx, "test/child/readme.md", strings.NewReader("child file"))
	source.WriteFile(ctx, "test/child/grand/note.md", strings.NewReader("grandchild file"))

	wrapper := ArchiveDecorator{Wrapped: source, Marker: ".archive"}

	content, err := wrapper.ReadFile(ctx, "test/.archive")
	assert.NoError(t, err)

	dest := NewMemoryStorage()
	wrapper.Wrapped = dest

	err = wrapper.WriteFile(ctx, "test/.archive", content)
	assert.NoError(t, err)

	assert.Equal(t, []byte("first file"), dest.Store["test/one.md"])
	assert.Equal(t, []byte("second file"), dest.Store["test/two.md"])
	assert.Equal(t, []byte("child file"), dest.Store["test/child/readme.md"])
	assert.Equal(t, []byte("grandchild file"), dest.Store["test/child/grand/note.md"])

}
