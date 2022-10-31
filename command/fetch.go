package command

import (
	"cas/tracing"
	"context"
	"fmt"

	"github.com/spf13/pflag"
)

type FetchCommand struct {
	Meta
}

func (c *FetchCommand) Name() string {
	return "Fetch"
}

func (c *FetchCommand) Synopsis() string {
	return "Fetches artifacts for a hash"
}

func (c *FetchCommand) Flags() *pflag.FlagSet {
	return pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)
}

func (c *FetchCommand) RunContext(ctx context.Context, args []string) error {

	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	if len(args) < 1 {
		return fmt.Errorf("this command takes at least 2 arguments: hash, and paths to upload")
	}

	hash := args[0]
	paths := args[1:]

	backend, err := c.createBackend(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	written, err := backend.FetchArtifacts(ctx, hash, paths)
	if err != nil {
		return tracing.Error(span, err)
	}

	c.print(hash)

	for _, value := range written {
		c.print(fmt.Sprintf("- %s", value))
	}

	return nil
}

// func (c *FetchCommand) write(ctx context.Context, path string, content io.Reader) error {
// 	ctx, span := c.tr.Start(ctx, "write")
// 	defer span.End()

// 	f, err := os.Create(path)
// 	if err != nil {
// 		return tracing.Error(span, err)
// 	}

// 	if _, err := io.Copy(f, content); err != nil {
// 		return tracing.Error(span, err)
// 	}

// 	return nil
// }
