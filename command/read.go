package command

import (
	"cas/tracing"
	"context"
	"fmt"

	"github.com/spf13/pflag"
)

type ReadCommand struct {
	Meta
}

func (c *ReadCommand) Name() string {
	return "read"
}

func (c *ReadCommand) Synopsis() string {
	return "Reads metadata for a hash"
}

func (c *ReadCommand) Flags() *pflag.FlagSet {
	return pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
}

func (c *ReadCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	if len(args) < 1 {
		return fmt.Errorf("this command takes at least 1 argument: hash, and optionally keys to read")
	}

	hash := args[0]
	keys := args[1:]

	backend, err := c.createBackend(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	meta, err := backend.ReadMetadata(ctx, hash, keys)
	if err != nil {
		return tracing.Error(span, err)
	}

	for key, value := range meta {
		c.print(fmt.Sprintf("%s: %s", key, value))
	}

	return nil
}
