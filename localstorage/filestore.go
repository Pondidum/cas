package localstorage

import (
	"cas/tracing"
	"context"
	"io"
	"os"
	"path"
	"sort"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var fsTrace = otel.Tracer("filestore")

type FileStore struct{}

func (m *FileStore) ListFiles(ctx context.Context, p string) ([]string, error) {
	ctx, span := fsTrace.Start(ctx, "list")
	defer span.End()

	files := []string{}

	if err := scanDir(ctx, &files, p); err != nil {
		return nil, tracing.Error(span, err)
	}

	sort.Strings(files)

	return files, nil
}

func scanDir(ctx context.Context, files *[]string, dirPath string) error {
	ctx, span := fsTrace.Start(ctx, "scanDir")
	defer span.End()

	span.SetAttributes(attribute.String("path", dirPath))

	contents, err := os.ReadDir(dirPath)
	if err != nil {
		return tracing.Error(span, err)
	}

	for _, file := range contents {

		if file.IsDir() {
			scanDir(ctx, files, path.Join(dirPath, file.Name()))

			continue
		}

		*files = append(*files, path.Join(dirPath, file.Name()))

	}

	return nil
}

func (fs *FileStore) ReadFile(ctx context.Context, p string) (io.ReadSeekCloser, error) {
	ctx, span := fsTrace.Start(ctx, "read")
	defer span.End()

	return os.Open(p)
}

func (fs *FileStore) WriteFile(ctx context.Context, p string, timestamp time.Time, content io.Reader) error {
	ctx, span := fsTrace.Start(ctx, "write")
	defer span.End()

	if err := os.MkdirAll(path.Dir(p), os.ModePerm); err != nil {
		return tracing.Error(span, err)
	}

	f, err := os.Create(p)
	if err != nil {
		return tracing.Error(span, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, content); err != nil {
		return tracing.Error(span, err)
	}

	if err := f.Close(); err != nil {
		return tracing.Error(span, err)
	}

	if err := os.Chtimes(p, timestamp, timestamp); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}
