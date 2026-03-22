package debug

import (
	"cas/config"
	"cas/hashing"
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

const HashesKey = "hashes"

func NewDebugger() *Debugger {
	return &Debugger{}
}

type Debugger struct {
	active bool
	root   string

	impl debug
}

func (d *Debugger) Flags() *config.ConfigGroup {
	flags := config.NewConfigGroup("debug")
	flags.BoolFlag(&d.active, "debug", "CAS_DEBUG", false, "enable storage of intermediate debug data")
	flags.StringFlag(&d.root, "debug-dir", "CAS_DEBUG_DIR", path.Join(os.TempDir(), "cas"), "where to temporarily store debug data")
	return flags
}

func (d *Debugger) Write(ctx context.Context, hash string, key string, content []byte) error {
	if d.impl == nil {
		if d.active {
			d.impl = &activeDebugger{root: path.Join(d.root, "@debug")}
		} else {
			d.impl = &nullDebugger{}
		}
	}

	return d.impl.Write(ctx, hash, key, content)
}
func (d *Debugger) All(ctx context.Context, hash string) (map[string]io.ReadSeekCloser, error) {
	if d.impl == nil {
		if d.active {
			d.impl = &activeDebugger{root: path.Join(d.root, "@debug")}
		} else {
			d.impl = &nullDebugger{}
		}
	}

	return d.impl.All(ctx, hash)
}

type debug interface {
	Write(ctx context.Context, hash string, key string, content []byte) error
	All(ctx context.Context, hash string) (map[string]io.ReadSeekCloser, error)
}

type nullDebugger struct{}

func (n *nullDebugger) Write(ctx context.Context, hash string, key string, content []byte) error {
	return nil
}

func (d *nullDebugger) All(ctx context.Context, hash string) (map[string]io.ReadSeekCloser, error) {
	return map[string]io.ReadSeekCloser{}, nil
}

type activeDebugger struct {
	root string
}

func (d *activeDebugger) Write(ctx context.Context, hash string, key string, content []byte) error {

	if err := os.MkdirAll(path.Join(d.root, hash), os.ModePerm); err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(d.root, hash, key), content, 0666); err != nil {
		return err
	}

	return nil
}

func (d *activeDebugger) All(ctx context.Context, hash string) (map[string]io.ReadSeekCloser, error) {

	files := map[string]io.ReadSeekCloser{}
	basePath := path.Join(d.root, hash)

	err := fs.WalkDir(os.DirFS(basePath), ".", func(relPath string, de fs.DirEntry, err error) error {
		if de.IsDir() {
			return nil
		}

		file, err := os.Open(path.Join(basePath, relPath))
		if err != nil {
			return err
		}
		files[relPath] = file
		return nil
	})

	return files, err
}

func MarshalIntermediates(hashes []hashing.FileHash) []byte {

	sb := strings.Builder{}
	for _, fh := range hashes {
		sb.WriteString(fh.Hash)
		sb.WriteString(" ")
		sb.WriteString(fh.Path)
		sb.WriteString("\n")
	}

	return []byte(sb.String())
}
