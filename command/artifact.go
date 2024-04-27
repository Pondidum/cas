package command

import (
	"cas/config"
	"cas/localstorage"
	"cas/tracing"
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
)

func NewArtifactCommand(storage localstorage.Storage) *ArtifactCommand {
	cmd := &ArtifactCommand{
		storage:    storage,
		backendCfg: NewBackendConfiguration(),
	}

	cmd.cfg = append(cmd.cfg, cmd.commandFlags())
	cmd.cfg = append(cmd.cfg, cmd.backendCfg.Flags()...)
	cmd.cfg = append(cmd.cfg, globalFlags())

	return cmd
}

type ArtifactCommand struct {
	cfg        []*config.ConfigGroup
	backendCfg *BackendConfiguration

	storage   localstorage.Storage
	statePath string
}

func (c *ArtifactCommand) Synopsis() string {
	return "Stores artifacts for a hash"
}

func (c *ArtifactCommand) Usages() []string {
	return []string{
		`cas artifact "${hash}" ./path/to/artifact`,
		`cas artifact "$<" "$@"`,
	}
}

func (c *ArtifactCommand) commandFlags() *config.ConfigGroup {
	cfg := config.NewConfigGroup("")

	cfg.StringFlag(&c.statePath, "state-path", "", ".cas/state", "the directory to hold local state")

	return cfg
}

func (c *ArtifactCommand) Configuration() []*config.ConfigGroup {
	return c.cfg
}

func (c *ArtifactCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := otel.Tracer("artifact").Start(ctx, "run")
	defer span.End()

	if len(args) < 2 {
		return fmt.Errorf("this command takes at least 2 arguments: hash, and paths to upload")
	}

	// we support receiving the hash directly, or the state file path
	// i.e. makefile using  `cas artifact "$<" some-file`)
	hash := strings.TrimPrefix(strings.TrimPrefix(args[0], c.statePath), "/")
	paths := args[1:]

	backend, err := c.backendCfg.Create(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	localFiles, err := localstorage.ReadMany(ctx, c.storage, paths)
	if err != nil {
		return tracing.Error(span, err)
	}

	written, err := backend.StoreArtifacts(ctx, hash, localFiles)
	if err != nil {
		return tracing.Error(span, err)
	}

	fmt.Fprintln(os.Stderr, "Storing artifacts for "+hash)

	for _, value := range written {
		fmt.Fprintf(os.Stderr, "- %s\n", value)
	}

	return nil
}
