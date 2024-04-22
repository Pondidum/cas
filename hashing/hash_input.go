package hashing

import (
	"cas/tracing"
	"context"
	"io"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tr = otel.Tracer("hashing")

type HashInputConfig struct {
	Algorithm string

	CliArgs []string

	// for Testing, either a final computed hash, or the input reader
	TestHash  string
	TestInput io.ReadCloser
}

func HashInput(ctx context.Context, conf HashInputConfig) (string, []string, error) {
	ctx, span := tr.Start(ctx, "hash_input")
	defer span.End()

	if conf.TestHash != "" {
		return conf.TestHash, []string{}, nil
	}

	input, err := selectInputSource(ctx, conf)
	if err != nil {
		return "", nil, tracing.Error(span, err)
	}
	defer input.Close()

	span.SetAttributes(attribute.String("hash_type", conf.Algorithm))
	hasher, err := NewHasher(conf.Algorithm)
	if err != nil {
		return "", nil, tracing.Error(span, err)
	}

	hash, intermediateHashes, err := hasher.Hash(ctx, input)
	if err != nil {
		return "", nil, tracing.Error(span, err)
	}

	span.SetAttributes(attribute.String("hash", hash))

	return hash, intermediateHashes, nil
}

func selectInputSource(ctx context.Context, conf HashInputConfig) (io.ReadCloser, error) {
	ctx, span := tr.Start(ctx, "select_input_source")
	defer span.End()

	if conf.TestInput != nil {
		span.SetAttributes(attribute.String("input_source", "test_input"))
		return conf.TestInput, nil
	}

	if len(conf.CliArgs) > 0 {
		inputFilePath := conf.CliArgs[0]

		span.SetAttributes(
			attribute.String("input_source", "file"),
			attribute.String("input_file", inputFilePath),
		)

		input, err := os.Open(inputFilePath)
		if err != nil {
			return nil, tracing.Error(span, err)
		}

		return input, nil
	}

	span.SetAttributes(attribute.String("input_source", "stdin"))

	return os.Stdin, nil
}
