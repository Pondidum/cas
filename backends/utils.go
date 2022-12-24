package backends

import (
	"cas/tracing"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
)

const MetadataTimeStamp = "@timestamp"

func ReadTimestamp(ctx context.Context, backend Backend, hash string) (time.Time, bool, error) {
	ctx, span := otel.Tracer("backends").Start(ctx, "read_timestamp")
	defer span.End()

	meta, err := backend.ReadMetadata(ctx, hash, []string{MetadataTimeStamp})
	if err != nil {
		return time.Time{}, false, tracing.Error(span, err)
	}

	timestampAnnotation, found := meta[MetadataTimeStamp]
	if !found {
		return time.Time{}, false, nil
	}

	seconds, err := strconv.ParseInt(timestampAnnotation, 10, 64)
	if err != nil {
		return time.Time{}, false, tracing.Error(span, err)
	}

	ts := time.Unix(seconds, 0)

	return ts, true, nil
}

func CreateHash(ctx context.Context, backend Backend, hash string, ts time.Time) error {
	ctx, span := otel.Tracer("backends").Start(ctx, "create_hash")
	defer span.End()

	stamp := strings.NewReader(fmt.Sprintf("%v", ts.Unix()))
	if err := backend.WriteMetadata(ctx, hash, MetadataTimeStamp, stamp); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}
