package command

import (
	"context"
	"os"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
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
