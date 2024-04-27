package config

import (
	"cas/tracing"
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tr = otel.Tracer("config")

type ConfigGroup struct {
	Name string

	flags       *pflag.FlagSet
	environment map[string]string
}

func NewConfigGroup(name string) *ConfigGroup {
	return &ConfigGroup{
		Name:        name,
		flags:       pflag.NewFlagSet(name, pflag.ContinueOnError),
		environment: map[string]string{},
	}
}

func (fg *ConfigGroup) StringFlag(target *string, flagName string, envVarName string, defaultValue string, usage string) {
	fg.flags.StringVar(target, flagName, defaultValue, usage)
	fg.environment[flagName] = envVarName
}

func (fg *ConfigGroup) Usages() []string {
	lines := []string{}

	fg.flags.VisitAll(func(flag *pflag.Flag) {

		usage := flag.Usage
		if flag.DefValue != "" {
			usage += fmt.Sprintf(" (default: %s)", flag.DefValue)
		}
		lines = append(lines, fmt.Sprintf("--%s|%s|%s", flag.Name, fg.environment[flag.Name], usage))
	})

	return lines
}

func (fg *ConfigGroup) Parse(ctx context.Context, args []string) error {
	ctx, span := tr.Start(ctx, "parse")
	defer span.End()

	if err := fg.flags.Parse(args); err != nil {
		return tracing.Error(span, err)
	}

	fg.flags.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			return
		}

		envVarName, found := fg.environment[f.Name]
		v := ""
		if found {
			v = os.Getenv(envVarName)
		}

		isDifferent := v != f.DefValue

		span.SetAttributes(
			attribute.Bool(f.Name+"_found", found),
			attribute.Bool(f.Name+"_different", isDifferent),
		)

		if found && v != "" && isDifferent {
			f.Value.Set(v)
		}
	})

	return nil
}

func (fg *ConfigGroup) Args() []string {
	return fg.flags.Args()
}
