package command

import (
	"cas/tracing"
	"context"
	"os"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func TestEnvironmentVariableOverrides(t *testing.T) {

	t.Run("no environment variables, no flags", func(t *testing.T) {
		os.Unsetenv("CAS_TEST_CHECK")
		os.Unsetenv(BackendEnvVar)

		ui := cli.NewMockUi()
		cmd := NewMockCommand(ui)
		assert.Equal(t, 0, cmd.Run([]string{}), ui.ErrorWriter.String())
		assert.Equal(t, "s3", cmd.backendName)
		assert.Equal(t, false, cmd.check)

	})

	t.Run("specified environment variables, no flags", func(t *testing.T) {

		os.Setenv("CAS_TEST_CHECK", "true")
		os.Setenv(BackendEnvVar, "testing")

		ui := cli.NewMockUi()
		cmd := NewMockCommand(ui)
		assert.Equal(t, 0, cmd.Run([]string{}), ui.ErrorWriter.String())
		assert.Equal(t, "testing", cmd.backendName)
		assert.Equal(t, true, cmd.check)
	})

	t.Run("specified environment variables, specified flags", func(t *testing.T) {

		os.Setenv("CAS_TEST_CHECK", "true")
		os.Setenv(BackendEnvVar, "testing")

		ui := cli.NewMockUi()
		cmd := NewMockCommand(ui)
		assert.Equal(t, 0, cmd.Run([]string{"--backend", "other", "--check=false"}), ui.ErrorWriter.String())
		assert.Equal(t, "other", cmd.backendName)
		assert.Equal(t, false, cmd.check)
	})
}

func TestTraceParent(t *testing.T) {

	exporter := tracing.NewMemoryExporter()

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSyncer(exporter),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	m := &Meta{
		tr:  tp.Tracer("mock"),
		cmd: &MockCommand{},
	}

	os.Setenv(TraceParentEnvVar, "00-7107538ee3f6bc77ada1b2d34a412e1d-bfe6177cefb76eb2-01")
	assert.Equal(t, 0, m.Run([]string{}))

	assert.Equal(t, "7107538ee3f6bc77ada1b2d34a412e1d", exporter.Spans[0].SpanContext().TraceID().String())
}

// --------------------------------------------------------------------------//

func NewMockCommand(ui cli.Ui) *MockCommand {
	cmd := &MockCommand{}
	cmd.Meta = NewMeta(ui, cmd)
	return cmd
}

type MockCommand struct {
	Meta

	check bool
}

func (c *MockCommand) Name() string {
	return "mock"
}
func (c *MockCommand) Synopsis() string {
	return "mock"
}
func (c *MockCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)

	flags.BoolVar(&c.check, "check", false, "")

	return flags
}
func (c *MockCommand) EnvironmentVariables() map[string]string {
	return map[string]string{
		"check": "CAS_TEST_CHECK",
	}
}

func (c *MockCommand) RunContext(ctx context.Context, args []string) error {
	return nil
}
