package command

import (
	"cas/tracing"
	"context"
	"fmt"

	"github.com/spf13/pflag"
)

type StoreCommand struct {
	Meta
}

func (c *StoreCommand) Name() string {
	return "store"
}

func (c *StoreCommand) Help() string {
	return ""
}

func (c *StoreCommand) Synopsis() string {
	return "Stores artifacts for a hash"
}

func (c *StoreCommand) Flags() *pflag.FlagSet {
	return pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
}

func (c *StoreCommand) RunContext(ctx context.Context, args []string) error {

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

	written, err := backend.StoreArtifacts(ctx, hash, paths)
	if err != nil {
		return tracing.Error(span, err)
	}

	c.print(hash)

	for _, value := range written {
		c.print(fmt.Sprintf("- %s", value))
	}

	return nil
}
