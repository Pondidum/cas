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

func NewArtifactListCommand(storage localstorage.Storage) *ArtifactListCommand {
	cmd := &ArtifactListCommand{
		storage:    storage,
		backendCfg: NewBackendConfiguration(),
	}

	cmd.cfg = append(cmd.cfg, cmd.commandFlags())
	cmd.cfg = append(cmd.cfg, cmd.backendCfg.Flags()...)
	cmd.cfg = append(cmd.cfg, globalFlags())

	return cmd
}

type ArtifactListCommand struct {
	cfg        []*config.ConfigGroup
	backendCfg *BackendConfiguration

	storage   localstorage.Storage
	statePath string
}

func (c *ArtifactListCommand) Synopsis() string {
	return "Lists artifacts for a hash"
}

func (c *ArtifactListCommand) Usages() []string {
	return []string{
		`cas artifact list "${hash}"`,
	}
}

func (c *ArtifactListCommand) commandFlags() *config.ConfigGroup {
	cfg := config.NewConfigGroup("")

	cfg.StringFlag(&c.statePath, "state-path", "", ".cas/state", "the directory to hold local state")

	return cfg
}

func (c *ArtifactListCommand) Configuration() []*config.ConfigGroup {
	return c.cfg
}

func (c *ArtifactListCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := otel.Tracer("artifact_list").Start(ctx, "run")
	defer span.End()

	if len(args) != 1 {
		return fmt.Errorf("this command takes exactly 1 argument: hash")
	}

	// we support receiving the hash directly, or the state file path
	// i.e. makefile using  `cas artifact "$<" some-file`)
	hash := strings.TrimPrefix(strings.TrimPrefix(args[0], c.statePath), "/")

	backend, err := c.backendCfg.Create(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	artifacts, err := backend.ListArtifacts(ctx, hash)
	if err != nil {
		return tracing.Error(span, err)
	}

	for _, artifact := range artifacts {
		fmt.Println(artifact)
	}

	return nil
}
