package command

import (
	"cas/tracing"
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type FetchCommand struct {
	Meta

	directory string
}

func (c *FetchCommand) Name() string {
	return "Fetch"
}

func (c *FetchCommand) Synopsis() string {
	return "Fetches artifacts for a hash"
}

func (c *FetchCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)

	pwd, _ := os.Getwd()
	flags.StringVar(&c.directory, "directory", pwd, "Where to write the artifacts to")

	return flags
}

func (c *FetchCommand) RunContext(ctx context.Context, args []string) error {

	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	if len(args) < 1 {
		return fmt.Errorf("this command takes at least 2 arguments: hash, and paths to upload")
	}

	hash := args[0]
	paths := args[1:]

	backend, err := c.createBackend(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	storage := c.createStorage(ctx)

	written, err := backend.FetchArtifacts(ctx, storage, hash, paths)
	if err != nil {
		return tracing.Error(span, err)
	}

	c.print(hash)

	for _, value := range written {
		c.print(fmt.Sprintf("- %s", value))
	}

	return nil
}
