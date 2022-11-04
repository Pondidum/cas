package command

import (
	"cas/version"
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/pflag"
)

type VersionCommand struct {
	Meta

	printLog bool
}

func (c *VersionCommand) Name() string {
	return "version"
}

func (c *VersionCommand) Help() string {
	return ""
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the version number"
}

func (c *VersionCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
	flags.BoolVar(&c.printLog, "changelog", false, "print the changelog")

	return flags
}

func (c *VersionCommand) RunContext(ctx context.Context, args []string) error {

	change := version.Changelog()
	c.Ui.Output(fmt.Sprintf(
		"%s - %s",
		change[0].Version,
		version.VersionNumber(),
	))

	if c.printLog {
		out, _ := glamour.Render(change[0].Log, "dark")
		c.Ui.Output(strings.TrimSpace(out))
	}

	return nil
}
