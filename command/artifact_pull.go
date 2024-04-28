package command

import (
	"cas/config"
	"cas/localstorage"
	"cas/tracing"
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
)

func NewArtifactPullCommand(storage localstorage.Storage) *ArtifactPullCommand {
	cmd := &ArtifactPullCommand{
		storage:    storage,
		backendCfg: NewBackendConfiguration(),
	}

	cmd.cfg = append(cmd.cfg, cmd.commandFlags())
	cmd.cfg = append(cmd.cfg, cmd.backendCfg.Flags()...)
	cmd.cfg = append(cmd.cfg, globalFlags())

	return cmd
}

type ArtifactPullCommand struct {
	cfg        []*config.ConfigGroup
	backendCfg *BackendConfiguration

	storage   localstorage.Storage
	statePath string
}

func (c *ArtifactPullCommand) Synopsis() string {
	return "Restores artifacts for a hash"
}

func (c *ArtifactPullCommand) Usages() []string {
	return []string{
		`cas artifact pull "${hash}"`,
		`cas artifact pull "${hash}" dist/index.js`,
	}
}

func (c *ArtifactPullCommand) commandFlags() *config.ConfigGroup {
	cfg := config.NewConfigGroup("")

	cfg.StringFlag(&c.statePath, "state-path", "", ".cas/state", "the directory to hold local state")

	return cfg
}

func (c *ArtifactPullCommand) Configuration() []*config.ConfigGroup {
	return c.cfg
}

func (c *ArtifactPullCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := otel.Tracer("artifact_pull").Start(ctx, "run")
	defer span.End()

	if len(args) < 1 {
		return fmt.Errorf("this command takes at least 1 argument: hash, and artifact paths to pull")
	}

	// we support receiving the hash directly, or the state file path
	// i.e. makefile using  `cas artifact "$<" some-file`)
	hash := strings.TrimPrefix(strings.TrimPrefix(args[0], c.statePath), "/")
	paths := args[1:]

	backend, err := c.backendCfg.Create(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	if len(paths) > 0 {
		for _, name := range paths {
			file, err := backend.FetchArtifact(ctx, hash, name)
			if err != nil {
				return tracing.Error(span, err)
			}
			defer file.Close()

			if err := c.storage.WriteFile(ctx, file.Name, file.Timestamp, file.Content); err != nil {
				return tracing.Error(span, err)
			}
		}
	} else {
		files, err := backend.FetchArtifacts(ctx, hash)
		if err != nil {
			return tracing.Error(span, err)
		}
		for _, file := range files {
			defer file.Close()

			if err := c.storage.WriteFile(ctx, file.Name, file.Timestamp, file.Content); err != nil {
				return tracing.Error(span, err)
			}
		}
	}

	return nil
}
