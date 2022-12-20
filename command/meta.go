package command

import (
	"cas/backends"
	"cas/backends/s3"
	"cas/localstorage"
	"cas/tracing"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/posener/complete"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Meta struct {
	Ui  cli.Ui
	cmd NamedCommand
	tr  trace.Tracer

	storage localstorage.Storage

	backendName  string
	outputFormat string
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
	EnvironmentVariables() map[string]string

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
	return m.cmd.Synopsis() + "\n\n" + m.allFlags().FlagUsages()
}

func (m *Meta) allFlags() *pflag.FlagSet {

	flags := m.cmd.Flags()

	defaultBackend := "s3"
	if v := os.Getenv("CAS_BACKEND"); v != "" {
		defaultBackend = v
	}

	flags.StringVar(&m.backendName, "backend", defaultBackend, "the backend to use for artifacts")
	flags.StringVar(&m.outputFormat, "output", "json", "The format to print results to the console with")

	return flags
}

func (m *Meta) applyEnvironmentFallback(ctx context.Context, flags *pflag.FlagSet) {
	ctx, span := m.tr.Start(ctx, "apply_environment_fallback")
	defer span.End()

	envVars := m.cmd.EnvironmentVariables()

	flags.VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			return
		}

		v, found := envVars[f.Name]
		isDifferent := v != f.DefValue

		span.SetAttributes(
			attribute.Bool(f.Name+"_found", found),
			attribute.Bool(f.Name+"_different", isDifferent),
		)

		if found && isDifferent {
			f.Value.Set(v)
		}
	})
}

func (m *Meta) Run(args []string) int {
	ctx := context.Background()

	ctx, span := m.tr.Start(ctx, m.cmd.Name())
	defer span.End()

	f := m.allFlags()

	if err := f.Parse(args); err != nil {
		tracing.Error(span, err)
		m.Ui.Error(err.Error())

		return 1
	}

	m.applyEnvironmentFallback(ctx, f)

	tracing.StoreFlags(ctx, f)

	if err := m.cmd.RunContext(ctx, f.Args()); err != nil {
		tracing.Error(span, err)
		m.Ui.Error(err.Error())

		return 1
	}

	return 0
}

func (m *Meta) createBackend(ctx context.Context) (backends.Backend, error) {
	ctx, span := m.tr.Start(ctx, "create_backend")
	defer span.End()

	span.SetAttributes(attribute.String("backend", m.backendName))

	switch strings.ToLower(m.backendName) {
	case "s3":
		cfg := s3.ConfigFromEnvironment()
		return s3.NewS3Backend(cfg), nil
	}

	return nil, fmt.Errorf("unsupported backend '%s'", m.backendName)
}

func (m *Meta) createStorage(ctx context.Context) localstorage.Storage {

	storage := m.storage
	if storage == nil {
		storage = &localstorage.FileStore{}
	}

	return localstorage.NewArchiveDecorator(storage)
}

func (m *Meta) print(line string) {
	m.Ui.Warn(line)
}
