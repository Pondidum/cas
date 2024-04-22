package localstorage

import (
	"cas/tracing"
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel"
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

func ReadMany(ctx context.Context, storage ReadableStorage, paths []string) ([]*LocalFile, error) {
	ctx, span := otel.Tracer("storage").Start(ctx, "read_many")
	defer span.End()

	files := make([]*LocalFile, 0, len(paths))

	for _, filePath := range paths {

		localFile, err := storage.ReadFile(ctx, filePath)
		if err != nil {
			closeAll(files)
			return nil, tracing.Error(span, err)
		}

		files = append(files, localFile)
	}

	return files, nil

}

func closeAll(files []*LocalFile) {
	for _, f := range files {
		f.Close()
	}
}
