package localstorage

import (
	"archive/tar"
	"bytes"
	"cas/tracing"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var archiveTrace = otel.Tracer("archive_decorator")

func NewArchiveDecorator(wrapped Storage) *ArchiveDecorator {
	return &ArchiveDecorator{
		Wrapped: wrapped,
		Marker:  ".archive",
	}
}

type ArchiveDecorator struct {
	Wrapped Storage

	Marker string
}

func (a *ArchiveDecorator) ListFiles(ctx context.Context, p string) ([]string, error) {
	// not sure on this implementation yet, so just pass through for now
	return a.Wrapped.ListFiles(ctx, p)
}

func (a *ArchiveDecorator) ReadFile(ctx context.Context, p string) (io.ReadSeekCloser, error) {
	ctx, span := archiveTrace.Start(ctx, "read")
	defer span.End()

	name := path.Base(p)
	isMarker := name == a.Marker

	span.SetAttributes(
		attribute.String("filename", name),
		attribute.String("marker", a.Marker),
		attribute.Bool("is_marker_file", isMarker),
	)

	if !isMarker {
		return a.Wrapped.ReadFile(ctx, p)
	}

	f, err := os.CreateTemp("", "cas-*.tar")
	if err != nil {
		return nil, tracing.Error(span, err)
	}
	defer f.Close()

	archive := tar.NewWriter(f)
	dirPath := path.Dir(p)

	span.SetAttributes(
		attribute.String("directory", dirPath),
		attribute.String("archive_file", f.Name()),
	)

	files, err := a.Wrapped.ListFiles(ctx, dirPath)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	span.SetAttributes(attribute.Int("file_found", len(files)))

	for _, file := range files {

		reader, err := a.Wrapped.ReadFile(ctx, file)
		if err != nil {
			return nil, tracing.Error(span, err)
		}

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, tracing.Error(span, err)
		}
		reader.Close()

		header := &tar.Header{
			Name:     strings.TrimPrefix(file, dirPath),
			ModTime:  time.Time{},
			Mode:     int64(0),
			Typeflag: tar.TypeReg,
			Size:     int64(len(content)), // fix this later to handle things bigger than int
		}

		if err := archive.WriteHeader(header); err != nil {
			return nil, tracing.Error(span, err)
		}

		if _, err := io.Copy(archive, bytes.NewReader(content)); err != nil {
			return nil, tracing.Error(span, err)
		}

	}

	if err := archive.Close(); err != nil {
		return nil, tracing.Error(span, err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		return nil, tracing.Error(span, err)
	}

	if err := a.Wrapped.WriteFile(ctx, p, time.Now(), f); err != nil {
		return nil, tracing.Error(span, err)
	}

	return a.Wrapped.ReadFile(ctx, p)
}

func (a *ArchiveDecorator) WriteFile(ctx context.Context, p string, timestamp time.Time, content io.Reader) error {
	ctx, span := archiveTrace.Start(ctx, "write")
	defer span.End()

	name := path.Base(p)

	if name != a.Marker {
		return a.Wrapped.WriteFile(ctx, p, timestamp, content)
	}

	archive := tar.NewReader(content)

	root := path.Dir(p)

	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return tracing.Error(span, err)
		}

		filepath := path.Join(root, header.Name)

		if err := a.Wrapped.WriteFile(ctx, filepath, timestamp, archive); err != nil {
			return err
		}
	}

	// we need to write a file with the right name too, so that make's existence check still passes
	if err := a.Wrapped.WriteFile(ctx, p, timestamp, &bytes.Buffer{}); err != nil {
		return err
	}

	return nil
}
