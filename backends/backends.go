package backends

import (
	"cas/localstorage"
	"context"
	"io"
)

type Backend interface {
	WriteMetadata(ctx context.Context, hash string, key string, value io.ReadSeeker) error
	ReadMetadata(ctx context.Context, hash string, keys []string) (map[string]string, error)

	StoreArtifacts(ctx context.Context, storage localstorage.ReadableStorage, hash string, paths []string) ([]string, error)
	FetchArtifacts(ctx context.Context, hash string, readFile ReadFile, writeFile WriteFile) error
}

type ReadFile func(ctx context.Context, relPath string) (io.ReadSeekCloser, error)
type WriteFile func(ctx context.Context, relPath string, content io.Reader) error
