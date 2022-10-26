package command

import (
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

func (c *ReadCommand) Help() string {
	return ""
}

func (c *ReadCommand) Synopsis() string {
	return "Reads metadata for a hash"
}

func (c *ReadCommand) Flags() *pflag.FlagSet {
	return pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
}

func (c *ReadCommand) RunContext(ctx context.Context, args []string) error {

	hash := args[0]

	keys := args[1:]

	c.Ui.Output(hash)

	for _, key := range keys {
		c.Ui.Info(fmt.Sprintf("%s: ...", key))
	}

	return nil
}
