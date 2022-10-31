package backends

import (
	"context"
	"io"
)

type Backend interface {
	WriteMetadata(ctx context.Context, hash string, data map[string]string) (map[string]string, error)
	ReadMetadata(ctx context.Context, hash string, keys []string) (map[string]string, error)

	StoreArtifacts(ctx context.Context, hash string, paths []string) ([]string, error)
	FetchArtifacts(ctx context.Context, hash string, paths []string) ([]string, error)
}

type Storage interface {
	ReadFile(ctx context.Context, p string) (io.ReadCloser, error)
	WriteFile(ctx context.Context, path string, content io.Reader) (string, error)
}
