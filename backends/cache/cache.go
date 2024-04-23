package cache

import (
	"cas/backends"
	"cas/localstorage"
	"cas/tracing"
	"context"
	"errors"
	"io"
	"os"
	"path"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("cache").Start(ctx, name)
}

func NewCachedBackend(wrapped backends.Backend) *CacheBackend {
	return &CacheBackend{
		wrapped: wrapped,
		root:    ".cas/cache",
	}
}

type CacheBackend struct {
	wrapped backends.Backend

	root string
}

// TODO: fix passthroughs

func (cache *CacheBackend) WriteMetadata(ctx context.Context, hash string, key string, value io.ReadSeeker) error {
	return cache.wrapped.WriteMetadata(ctx, hash, key, value)
}

func (cache *CacheBackend) ReadMetadata(ctx context.Context, hash string, keys []string) (map[string]string, error) {
	return cache.wrapped.ReadMetadata(ctx, hash, keys)
}

func (cache *CacheBackend) StoreArtifacts(ctx context.Context, hash string, files []*localstorage.LocalFile) ([]string, error) {
	return cache.wrapped.StoreArtifacts(ctx, hash, files)
}

func (cache *CacheBackend) ListArtifacts(ctx context.Context, hash string) ([]string, error) {
	// this is always pass through, as we don't know if new artifacts have been
	// added to a hash after our last list/read call
	return cache.wrapped.ListArtifacts(ctx, hash)
}

func (cache *CacheBackend) FetchArtifact(ctx context.Context, hash string, name string) (*backends.RemoteFile, error) {
	ctx, span := startSpan(ctx, "fetch_artifact")
	defer span.End()

	file, err := cache.readCacheFile(ctx, hash, name)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	if file != nil {
		return file, nil
	}

	remoteFile, err := cache.wrapped.FetchArtifact(ctx, hash, name)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	if err := cache.writeCacheFile(ctx, hash, remoteFile); err != nil {
		return nil, tracing.Error(span, err)
	}
	remoteFile.Close()

	// now pull back from the local cache
	file, err = cache.readCacheFile(ctx, hash, name)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	return file, nil
}

func (cache *CacheBackend) FetchArtifacts(ctx context.Context, hash string) ([]*backends.RemoteFile, error) {
	ctx, span := startSpan(ctx, "fetch_artifacts")
	defer span.End()

	paths, err := cache.wrapped.ListArtifacts(ctx, hash)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	remoteFiles := make([]*backends.RemoteFile, 0, len(paths))
	for _, path := range paths {
		remoteFile, err := cache.FetchArtifact(ctx, hash, path)
		if err != nil {
			closeAll(remoteFiles)
			return nil, tracing.Error(span, err)
		}

		remoteFiles = append(remoteFiles, remoteFile)
	}

	return remoteFiles, nil
}

func (cache *CacheBackend) readCacheFile(ctx context.Context, hash string, name string) (*backends.RemoteFile, error) {
	ctx, span := startSpan(ctx, "read_cache_file")
	defer span.End()

	cachePath := cache.cachePathFor(hash, name)
	file, err := os.Open(cachePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, tracing.Error(span, err)
	}

	stats, err := file.Stat()
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	return &backends.RemoteFile{
		Name:      name,
		Timestamp: stats.ModTime(),
		Content:   file,
	}, nil
}

func (cache *CacheBackend) writeCacheFile(ctx context.Context, hash string, remoteFile *backends.RemoteFile) error {
	ctx, span := startSpan(ctx, "write_cache_file")
	defer span.End()

	localPath := cache.cachePathFor(hash, remoteFile.Name)

	if err := os.MkdirAll(path.Dir(localPath), os.ModePerm); err != nil {
		return tracing.Error(span, err)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return tracing.Error(span, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, remoteFile.Content); err != nil {
		return tracing.Error(span, err)
	}

	if err := f.Close(); err != nil {
		return tracing.Error(span, err)
	}

	if err := os.Chtimes(localPath, remoteFile.Timestamp, remoteFile.Timestamp); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}

func (cache *CacheBackend) cachePathFor(hash string, name string) string {
	return path.Join(cache.root, hash, name)
}

func closeAll(files []*backends.RemoteFile) {
	for _, f := range files {
		f.Close()
	}
}
