package command

import (
	"cas/config"
	"cas/version"
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

func NewVersionCommand() *VersionCommand {
	cmd := &VersionCommand{}

	cmd.cfg = append(cmd.cfg, cmd.commandFlags())
	cmd.cfg = append(cmd.cfg, globalFlags())

	return cmd
}

type VersionCommand struct {
	cfg []*config.ConfigGroup

	printLog bool
	rawLog   bool
	short    bool
}

func (c *VersionCommand) Name() string {
	return "version"
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the version number"
}

func (c *VersionCommand) Usages() []string {
	return []string{
		`cas artifact "${hash}" ./path/to/artifact`,
		`cas artifact "$<" "$@"`,
	}
}

func (c *VersionCommand) commandFlags() *config.ConfigGroup {
	cfg := config.NewConfigGroup("")

	cfg.BoolFlag(&c.printLog, "changelog", "", false, "print the changelog")
	cfg.BoolFlag(&c.short, "short", "", false, "show only the version, not the sha")
	cfg.BoolFlag(&c.rawLog, "raw", "", false, "print the plain markdown from the changelog")

	return cfg
}

func (c *VersionCommand) Configuration() []*config.ConfigGroup {
	return c.cfg
}

func (c *VersionCommand) RunContext(ctx context.Context, args []string) error {

	change := version.Changelog()

	if c.short {
		fmt.Println(change[0].Version)
	} else {
		fmt.Println(fmt.Sprintf(
			"%s - %s",
			change[0].Version,
			version.VersionNumber(),
		))
	}

	if c.printLog {
		if c.rawLog {
			fmt.Println(strings.TrimSpace(change[0].Log))
		} else {
			out, _ := glamour.Render(change[0].Log, "dark")
			fmt.Println(strings.TrimSpace(out))
		}
	}

	return nil
}
