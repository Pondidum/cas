package command

import (
	"bytes"
	"cas/backends"
	"cas/hashing"
	"cas/localstorage"
	"cas/tracing"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

func NewFetchCommand(ui cli.Ui, storage localstorage.Storage) *FetchCommand {
	cmd := &FetchCommand{storage: storage}
	cmd.Meta = NewMeta(ui, cmd)
	return cmd
}

type FetchCommand struct {
	Meta

	storage localstorage.Storage

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
		"verbose": "CAS_VERBOSE",
		"debug":   "CAS_DEBUG",
	}

}

func (c *FetchCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	hash, intermediate, err := hashing.HashInput(ctx, hashing.HashInputConfig{
		Algorithm: c.algorithm,
		CliArgs:   args,
		TestHash:  c.testHash,
		TestInput: c.testInput,
	})
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

	statePath := path.Join(c.statePath, hash)
	if err := c.storage.WriteFile(ctx, statePath, ts, &bytes.Buffer{}); err != nil {
		return tracing.Error(span, err)
	}

	remoteFiles, err := backend.FetchArtifacts(ctx, hash)
	if err != nil {
		return tracing.Error(span, err)
	}

	for _, remoteFile := range remoteFiles {
		c.verbosePrint("Fetching artifact: " + remoteFile.Name)

		if err := c.storage.WriteFile(ctx, remoteFile.Name, remoteFile.Timestamp, remoteFile.Content); err != nil {
			return tracing.Error(span, err)
		}

		remoteFile.Close()
	}

	c.Ui.Output(statePath)

	return nil
}

func (c *FetchCommand) verbosePrint(line string) {
	if c.verbose {
		c.print(line)
	}
}
