package command

import (
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

	hash := args[0]

	paths := args[1:]

	c.Ui.Output(hash)

	for _, key := range paths {
		c.Ui.Info(fmt.Sprintf("- %s", key))
	}

	return nil
}
