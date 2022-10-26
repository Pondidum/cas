package command

import (
	"cas/tracing"
	"context"
	"fmt"

	"github.com/spf13/pflag"
)

type WriteCommand struct {
	Meta
}

func (c *WriteCommand) Name() string {
	return "write"
}

func (c *WriteCommand) Help() string {
	return ""
}

func (c *WriteCommand) Synopsis() string {
	return "Writes metadata for a hash"
}

func (c *WriteCommand) Flags() *pflag.FlagSet {
	return pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
}

func (c *WriteCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	if len(args) < 1 {
		return fmt.Errorf("this command takes at least 1 argument: hash, and optionally key=value pairs")
	}

	hash := args[0]

	data, err := parseKeyValuePairs(args[1:])
	if err != nil {
		return tracing.Error(span, err)
	}

	c.Ui.Output(hash)

	for key, value := range data {
		c.Ui.Info(fmt.Sprintf("%s: %s", key, value))
	}

	return nil
}
