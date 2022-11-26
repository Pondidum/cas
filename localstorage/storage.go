package localstorage

import (
	"context"
	"io"
)

type ReadableStorage interface {
	ReadFile(ctx context.Context, p string) (io.ReadCloser, error)
}

type WritableStorage interface {
	WriteFile(ctx context.Context, path string, content io.Reader) (string, error)
}

type Storage interface {
	ReadableStorage
	WritableStorage
}
