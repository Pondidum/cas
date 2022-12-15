package backends

import (
	"cas/tracing"
	"context"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
)

const MetadataTimeStamp = "@timestamp"

func ReadTimestamp(ctx context.Context, backend Backend, hash string) (*time.Time, error) {
	ctx, span := otel.Tracer("backends").Start(ctx, "read_timestamp")
	defer span.End()

	meta, err := backend.ReadMetadata(ctx, hash, []string{MetadataTimeStamp})
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	timestampAnnotation, found := meta[MetadataTimeStamp]
	if !found {
		return nil, nil
	}

	seconds, err := strconv.ParseInt(timestampAnnotation, 10, 64)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	ts := time.Unix(seconds, 0)

	return &ts, nil
}
