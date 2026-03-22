package debug

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWritingAndReading(t *testing.T) {

	root, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(root)

	d := &activeDebugger{root: root}
	require.NoError(t, d.Write(t.Context(), "aaaa", "hashes", []byte("content")))

	all, err := d.All(t.Context(), "aaaa")
	require.NoError(t, err)

	require.Len(t, all, 1)
	require.Contains(t, all, "hashes")
	content, _ := io.ReadAll(all["hashes"])
	require.Equal(t, "content", string(content))
}
