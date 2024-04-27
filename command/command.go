package command

import (
	"cas/config"
	"cas/tracing"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/ryanuber/columnize"
	"go.opentelemetry.io/otel"
)

type CommandDefinition interface {
	Synopsis() string
	Usages() []string
	Configuration() []*config.ConfigGroup
	RunContext(ctx context.Context, args []string) error
}

func NewCommand(name string, definition CommandDefinition) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return &command{definition, name}, nil
	}
}

type command struct {
	CommandDefinition

	name string
}

func (c *command) Help() string {
	sb := strings.Builder{}

	sb.WriteString(c.Synopsis())
	sb.WriteString("\n\n")

	usages := c.Usages()
	if len(usages) > 0 {
		sb.WriteString("Usage:\n\n")
		for _, usage := range usages {
			sb.WriteString("      ")
			sb.WriteString(usage)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	lines := []string{}
	for _, group := range c.Configuration() {
		if group.Name != "" {
			lines = append(lines, strings.Title(group.Name)+" Flags:|EnvVar|Usage")
		} else {
			lines = append(lines, "Command"+" Flags:|EnvVar|Usage")
		}

		lines = append(lines, "")
		lines = append(lines, group.Usages()...)
		lines = append(lines, "")
	}

	sb.WriteString(columnize.Format(lines, &columnize.Config{
		Glue:   "    ",
		Prefix: "",
	}))

	return sb.String()
}

func (c *command) Run(args []string) int {
	// note: traceParent is read here rather than with the flags, as we need the value available
	// before we start parsing flags/etc.
	ctx := tracing.WithTraceParent(context.Background(), os.Getenv(TraceParentEnvVar))

	ctx, span := otel.Tracer("command").Start(ctx, c.name)
	defer span.End()

	f := config.Combine(c.name, c.Configuration()...)

	if err := f.Parse(ctx, args); err != nil {
		tracing.Error(span, err)
		fmt.Fprintln(os.Stderr, err.Error())

		return 1
	}

	if err := c.RunContext(ctx, f.Args()); err != nil {
		tracing.Error(span, err)
		fmt.Fprintln(os.Stderr, err.Error())

		return 1
	}

	return 0
}
