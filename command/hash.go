package command

import (
	"cas/config"
	"cas/hashing"
	"cas/tracing"
	"context"
	"fmt"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func NewHashCommand() *HashCommand {
	cmd := &HashCommand{}

	cmd.cfg = append(cmd.cfg, cmd.commandFlags())
	cmd.cfg = append(cmd.cfg, globalFlags())

	return cmd
}

type HashCommand struct {
	cfg []*config.ConfigGroup

	algorithm string

	// for testing hashing on streams of data
	testInput io.ReadCloser
}

func (c *HashCommand) Synopsis() string {
	return "Hashes files"
}

func (c *HashCommand) Usages() []string {
	return []string{
		`cas hash file-list.txt`,
		`find . -type f | cas hash`,
	}
}

func (c *HashCommand) commandFlags() *config.ConfigGroup {
	cfg := config.NewConfigGroup("")

	cfg.StringFlag(&c.algorithm, "algorithm", "", "sha256", "change the hashing algorithm used")

	return cfg
}

func (c *HashCommand) Configuration() []*config.ConfigGroup {
	return c.cfg
}

func (c *HashCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := otel.Tracer("hash").Start(ctx, "run")
	defer span.End()

	hash, _, err := hashing.HashInput(ctx, hashing.HashInputConfig{
		Algorithm: c.algorithm,
		CliArgs:   args,
		TestInput: c.testInput,
	})
	if err != nil {
		return tracing.Error(span, err)
	}

	span.SetAttributes(attribute.String("hash", hash))

	// hash is expected to be used in shell scripts, rather than in the makefile's rule area, so
	// instead we output to stdout so that `hash=$(find .... | cas hash)` works as expected.
	fmt.Println(hash)

	return nil
}
