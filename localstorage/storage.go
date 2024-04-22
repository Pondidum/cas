package localstorage

import (
	"context"
	"io"
	"time"
)

type ReadableStorage interface {
	ListFiles(ctx context.Context, p string) ([]string, error)
	ReadFile(ctx context.Context, p string) (*LocalFile, error)
}

type WritableStorage interface {
	WriteFile(ctx context.Context, path string, timestamp time.Time, content io.Reader) error
}

type Storage interface {
	ReadableStorage
	WritableStorage
}

type LocalFile struct {
	Path    string
	Content io.ReadSeekCloser
}

func (lf *LocalFile) Close() error {
	if lf.Content != nil {
		return lf.Content.Close()
	}
	return nil
}
