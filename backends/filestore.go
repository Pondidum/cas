package backends

import (
	"cas/tracing"
	"context"
	"io"
	"os"

	"go.opentelemetry.io/otel"
)

var fsTrace = otel.Tracer("filestore")

type FileStore struct{}

func (fs *FileStore) ReadFile(ctx context.Context, p string) (io.ReadCloser, error) {
	ctx, span := fsTrace.Start(ctx, "read")
	defer span.End()

	return os.Open(p)
}

func (fs *FileStore) WriteFile(ctx context.Context, path string, content io.Reader) (string, error) {
	ctx, span := fsTrace.Start(ctx, "write")
	defer span.End()

	f, err := os.Create(path)
	if err != nil {
		return "", tracing.Error(span, err)
	}

	if _, err := io.Copy(f, content); err != nil {
		return "", tracing.Error(span, err)
	}

	return path, nil
}
