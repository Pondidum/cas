package command

import (
	"bytes"
	"cas/backends"
	"cas/hashing"
	"cas/tracing"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

func NewFetchCommand(ui cli.Ui) *FetchCommand {
	cmd := &FetchCommand{}
	cmd.Meta = NewMeta(ui, cmd)
	return cmd
}

type FetchCommand struct {
	Meta

	algorithm string
	statePath string
	verbose   bool
	debug     bool

	// for testing hashing on streams of data
	testInput io.ReadCloser

	// for skipping hashing to use a specific value
	testHash string
}

func (c *FetchCommand) Name() string {
	return "Fetch"
}

func (c *FetchCommand) Synopsis() string {
	return "Fetches state and artifacts for a set of files"
}

func (c *FetchCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)

	flags.StringVar(&c.statePath, "state-path", ".cas/state", "the directory to hold local state")
	flags.StringVar(&c.algorithm, "algorithm", "sha256", "change the hashing algorithm used")
	flags.BoolVar(&c.verbose, "verbose", false, "print more information")
	flags.BoolVar(&c.debug, "debug", false, "write a debug file to the backing store")

	return flags
}

func (c *FetchCommand) EnvironmentVariables() map[string]string {

	return map[string]string{
		"verbose": os.Getenv("CAS_VERBOSE"),
		"debug":   os.Getenv("CAS_DEBUG"),
	}

}

func (c *FetchCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	hash, intermediate, err := c.hashInput(ctx, args)
	if err != nil {
		return tracing.Error(span, err)
	}

	c.verbosePrint(fmt.Sprintf("Hash: %s", hash))

	backend, err := c.createBackend(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	ts, timestampExists, err := backends.ReadTimestamp(ctx, backend, hash)
	if err != nil {
		return tracing.Error(span, err)
	}

	span.SetAttributes(attribute.Bool("existing_hash", timestampExists))
	if !timestampExists {
		ts = time.Now()

		if err := backends.CreateHash(ctx, backend, hash, ts); err != nil {
			return tracing.Error(span, err)
		}
	}

	if c.debug {
		backend.WriteMetadata(ctx, hash, "@debug/hashes", strings.NewReader(strings.Join(intermediate, "")))
	}

	storage := c.createStorage(ctx)

	statePath := path.Join(c.statePath, hash)
	if err := storage.WriteFile(ctx, statePath, ts, &bytes.Buffer{}); err != nil {
		return tracing.Error(span, err)
	}

	writeArtifact := func(ctx context.Context, relPath string, content io.Reader) error {
		c.verbosePrint("Fetching artifact: " + relPath)
		return storage.WriteFile(ctx, relPath, ts, content)
	}

	if err := backend.FetchArtifacts(ctx, hash, storage.ReadFile, writeArtifact); err != nil {
		return tracing.Error(span, err)
	}

	c.Ui.Output(statePath)

	return nil
}

func (c *FetchCommand) selectInputSource(ctx context.Context, args []string) (io.ReadCloser, error) {
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

func (c *FetchCommand) hashInput(ctx context.Context, args []string) (string, []string, error) {
	ctx, span := c.tr.Start(ctx, "hash_input")
	defer span.End()

	if c.testHash != "" {
		return c.testHash, []string{}, nil
	}

	input, err := c.selectInputSource(ctx, args)
	if err != nil {
		return "", nil, tracing.Error(span, err)
	}
	defer input.Close()

	span.SetAttributes(attribute.String("hash_type", c.algorithm))
	hasher, err := hashing.NewHasher(c.algorithm)
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

func (c *FetchCommand) verbosePrint(line string) {
	if c.verbose {
		c.print(line)
	}
}
