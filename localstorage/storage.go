package localstorage

import (
	"context"
	"io"
)

type ReadableStorage interface {
	ListFiles(ctx context.Context, p string) ([]string, error)
	ReadFile(ctx context.Context, p string) (io.ReadCloser, error)
}

type WritableStorage interface {
	WriteFile(ctx context.Context, path string, content io.Reader) error
}

type Storage interface {
	ReadableStorage
	WritableStorage
}
