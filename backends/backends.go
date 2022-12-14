package backends

import (
	"cas/localstorage"
	"context"
	"io"
)

type Backend interface {
	WriteMetadata(ctx context.Context, hash string, data map[string]string) (map[string]string, error)
	ReadMetadata(ctx context.Context, hash string, keys []string) (map[string]string, error)

	StoreArtifacts(ctx context.Context, storage localstorage.ReadableStorage, hash string, paths []string) ([]string, error)
	FetchArtifacts(ctx context.Context, hash string, writeFile WriteFile) error
}

type WriteFile func(ctx context.Context, relPath string, content io.Reader) error
