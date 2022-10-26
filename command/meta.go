package command

import (
	"cas/tracing"
	"context"

	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Meta struct {
	Ui cli.Ui

	cmd NamedCommand
	tr  trace.Tracer
}

func NewMeta(ui cli.Ui, cmd NamedCommand) Meta {
	return Meta{
		Ui:  ui,
		cmd: cmd,
		tr:  otel.Tracer(cmd.Name()),
	}
}

type NamedCommand interface {
	Name() string
	Synopsis() string

	Flags() *pflag.FlagSet
	RunContext(ctx context.Context, args []string) error
}

func (m *Meta) AutocompleteFlags() complete.Flags {
	// return m.cmd.Flags().Autocomplete()
	return nil
}

func (m *Meta) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (m *Meta) Help() string {
	return m.cmd.Synopsis() + "\n\n" + m.cmd.Flags().FlagUsages()
}

func (m *Meta) Run(args []string) int {
	ctx := context.Background()

	ctx, span := m.tr.Start(ctx, m.cmd.Name())
	defer span.End()

	f := m.cmd.Flags()

	if err := f.Parse(args); err != nil {
		tracing.Error(span, err)
		m.Ui.Error(err.Error())

		return 1
	}

	tracing.StoreFlags(ctx, f)

	if err := m.cmd.RunContext(ctx, f.Args()); err != nil {
		tracing.Error(span, err)
		m.Ui.Error(err.Error())

		return 1
	}

	return 0
}
