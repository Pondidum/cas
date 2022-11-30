package localstorage

import (
	"cas/tracing"
	"context"
	"io"
	"os"
	"path"

	"go.opentelemetry.io/otel"
)

var fsTrace = otel.Tracer("filestore")

type FileStore struct{}

func (fs *FileStore) ReadFile(ctx context.Context, p string) (io.ReadCloser, error) {
	ctx, span := fsTrace.Start(ctx, "read")
	defer span.End()

	return os.Open(p)
}

func (fs *FileStore) WriteFile(ctx context.Context, p string, content io.Reader) error {
	ctx, span := fsTrace.Start(ctx, "write")
	defer span.End()

	if err := os.MkdirAll(path.Dir(p), os.ModePerm); err != nil {
		return tracing.Error(span, err)
	}

	f, err := os.Create(p)
	if err != nil {
		return tracing.Error(span, err)
	}

	if _, err := io.Copy(f, content); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}
