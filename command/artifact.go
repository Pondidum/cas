package command

import (
	"cas/localstorage"
	"cas/tracing"
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
)

func NewArtifactCommand(ui cli.Ui, storage localstorage.Storage) *ArtifactCommand {
	cmd := &ArtifactCommand{storage: storage}
	cmd.Meta = NewMeta(ui, cmd)
	return cmd
}

type ArtifactCommand struct {
	Meta

	storage   localstorage.Storage
	statePath string
}

func (c *ArtifactCommand) Name() string {
	return "artifact"
}

func (c *ArtifactCommand) Synopsis() string {
	return "Stores artifacts for a hash"
}

func (c *ArtifactCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)

	flags.StringVar(&c.statePath, "state-path", ".cas/state", "the directory to hold local state")

	return flags
}

func (c *ArtifactCommand) EnvironmentVariables() map[string]string {
	return map[string]string{}
}

func (c *ArtifactCommand) RunContext(ctx context.Context, args []string) error {

	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	if len(args) < 2 {
		return fmt.Errorf("this command takes at least 2 arguments: hash, and paths to upload")
	}

	// we support receiving the hash directly, or the state file path
	// i.e. makefile using  `cas artifact "$<" some-file`)
	hash := strings.TrimPrefix(strings.TrimPrefix(args[0], c.statePath), "/")
	paths := args[1:]

	backend, err := c.createBackend(ctx)
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

	c.print("Storing artifacts for " + hash)

	for _, value := range written {
		c.print(fmt.Sprintf("- %s", value))
	}

	return nil
}
