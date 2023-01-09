package command

import (
	"cas/hashing"
	"cas/tracing"
	"context"
	"io"
	"os"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

func NewHashCommand(ui cli.Ui) *HashCommand {
	cmd := &HashCommand{}
	cmd.Meta = NewMeta(ui, cmd)
	return cmd
}

type HashCommand struct {
	Meta

	algorithm string

	// for testing hashing on streams of data
	testInput io.ReadCloser
}

func (c *HashCommand) Name() string {
	return "Fetch"
}

func (c *HashCommand) Synopsis() string {
	return "Fetches state and artifacts for a set of files"
}

func (c *HashCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)

	flags.StringVar(&c.algorithm, "algorithm", "sha256", "change the hashing algorithm used")

	return flags
}

func (c *HashCommand) EnvironmentVariables() map[string]string {
	return map[string]string{}
}

func (c *HashCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	input, err := c.selectInputSource(ctx, args)
	if err != nil {
		return tracing.Error(span, err)
	}
	defer input.Close()

	span.SetAttributes(attribute.String("hash_type", c.algorithm))
	hasher, err := hashing.NewHasher(c.algorithm)
	if err != nil {
		return tracing.Error(span, err)
	}

	hash, _, err := hasher.Hash(ctx, input)
	if err != nil {
		return tracing.Error(span, err)
	}

	span.SetAttributes(attribute.String("hash", hash))

	// hash is expected to be used in shell scripts, rather than in the makefile's rule area, so
	// instead we output to stdout so that `hash=$(find .... | cas hash)` works as expected.
	c.Ui.Output(hash)

	return nil
}

func (c *HashCommand) selectInputSource(ctx context.Context, args []string) (io.ReadCloser, error) {
	ctx, span := c.tr.Start(ctx, "select_input_source")
	defer span.End()

	if c.testInput != nil {
		span.SetAttributes(attribute.String("input_source", "test_input"))
		return c.testInput, nil
	}

	if len(args) > 0 {
		inputFilePath := args[0]

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
