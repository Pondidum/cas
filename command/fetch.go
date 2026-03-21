package command

import (
	"bytes"
	"cas/backends"
	"cas/config"
	"cas/hashing"
	"cas/localstorage"
	"cas/tracing"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func NewFetchCommand(storage localstorage.Storage) *FetchCommand {
	cmd := &FetchCommand{
		storage:    storage,
		backendCfg: NewBackendConfiguration(),
	}

	cmd.cfg = append(cmd.cfg, cmd.commandFlags())
	cmd.cfg = append(cmd.cfg, cmd.backendCfg.Flags()...)
	cmd.cfg = append(cmd.cfg, globalFlags())

	// cmd.Meta = NewMeta(ui, cmd)
	return cmd
}

type FetchCommand struct {
	cfg        []*config.ConfigGroup
	backendCfg *BackendConfiguration

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

func (c *FetchCommand) Synopsis() string {
	return "Fetches state and artifacts for a set of files for use in a makefile rule"
}

func (c *FetchCommand) Usages() []string {
	return []string{
		`find . -type=f | cas fetch`,
		`cas fetch file-list.txt`,
	}
}
func (c *FetchCommand) commandFlags() *config.ConfigGroup {
	cfg := config.NewConfigGroup("")

	cfg.StringFlag(&c.statePath, "state-path", "", ".cas/state", "the directory to hold local state")
	cfg.StringFlag(&c.algorithm, "algorithm", "", "sha256", "change the hashing algorithm used")
	cfg.BoolFlag(&c.verbose, "verbose", "CAS_VERBOSE", false, "print more information")
	cfg.BoolFlag(&c.debug, "debug", "CAS_DEBUG", false, "write a debug file to the backing store")

	return cfg
}

func (c *FetchCommand) Configuration() []*config.ConfigGroup {
	return c.cfg
}

func (c *FetchCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := otel.Tracer("fetch").Start(ctx, "run")
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

	backend, err := c.backendCfg.Create(ctx)
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
		sb := &strings.Builder{}
		for _, fh := range intermediate {
			sb.WriteString(fh.Hash)
			sb.WriteString(" ")
			sb.WriteString(fh.Path)
			sb.WriteString("\n")
		}

		backend.WriteMetadata(ctx, hash, "@debug/hashes", strings.NewReader(sb.String()))
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

	fmt.Println(statePath)

	return nil
}

func (c *FetchCommand) verbosePrint(line string) {
	if c.verbose {
		fmt.Fprintln(os.Stderr, line)
	}
}
