package command

import (
	"cas/config"
	"cas/tracing"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func TestEnvironmentVariableOverrides(t *testing.T) {

	t.Run("no environment variables, no flags", func(t *testing.T) {
		os.Unsetenv("CAS_TEST_CHECK")
		os.Unsetenv(BackendEnvVar)

		cmd := NewMockCommand()
		wrapped, _ := NewCommand("mock-flags", cmd)()

		assert.Equal(t, 0, wrapped.Run([]string{}))
		assert.Equal(t, "s3", cmd.backendConfig.name)
		assert.Equal(t, false, cmd.check)

	})

	t.Run("specified environment variables, no flags", func(t *testing.T) {

		os.Setenv("CAS_TEST_CHECK", "true")
		os.Setenv(BackendEnvVar, "testing")

		cmd := NewMockCommand()
		wrapped, _ := NewCommand("mock-flags", cmd)()

		assert.Equal(t, 0, wrapped.Run([]string{}))
		assert.Equal(t, "testing", cmd.backendConfig.name)
		assert.Equal(t, true, cmd.check)
	})

	t.Run("specified environment variables, specified flags", func(t *testing.T) {

		os.Setenv("CAS_TEST_CHECK", "true")
		os.Setenv(BackendEnvVar, "testing")

		cmd := NewMockCommand()
		wrapped, _ := NewCommand("mock-flags", cmd)()

		assert.Equal(t, 0, wrapped.Run([]string{"--backend", "other", "--check=false"}))
		assert.Equal(t, "other", cmd.backendConfig.name)
		assert.Equal(t, false, cmd.check)
	})
}

func TestTraceParent(t *testing.T) {

	exporter := tracing.NewMemoryExporter()

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSyncer(exporter),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	cmd := &command{
		CommandDefinition: NewMockCommand(),
		name:              "mock",
		tracer:            tp.Tracer("mock"),
	}

	os.Setenv(TraceParentEnvVar, "00-7107538ee3f6bc77ada1b2d34a412e1d-bfe6177cefb76eb2-01")
	assert.Equal(t, 0, cmd.Run([]string{}))

	assert.Equal(t, "7107538ee3f6bc77ada1b2d34a412e1d", exporter.Spans[0].SpanContext().TraceID().String())
}

// --------------------------------------------------------------------------//

func NewMockCommand() *MockCommand {
	cmd := &MockCommand{
		backendConfig: NewBackendConfiguration(),
	}

	cmd.cfg = append(cmd.cfg, cmd.commandFlags())
	cmd.cfg = append(cmd.cfg, cmd.backendConfig.Flags()...)
	return cmd
}

type MockCommand struct {
	cfg           []*config.ConfigGroup
	backendConfig *BackendConfiguration
	check         bool
}

func (c *MockCommand) Synopsis() string {
	return "mock"
}

func (c *MockCommand) Usages() []string {
	return []string{
		`cas artifact "${hash}" ./path/to/artifact`,
		`cas artifact "$<" "$@"`,
	}
}

func (c *MockCommand) commandFlags() *config.ConfigGroup {
	cfg := config.NewConfigGroup("")

	cfg.BoolFlag(&c.check, "check", "CAS_TEST_CHECK", false, "")

	return cfg
}

func (c *MockCommand) Configuration() []*config.ConfigGroup {
	return c.cfg
}

func (c *MockCommand) RunContext(ctx context.Context, args []string) error {
	return nil
}
