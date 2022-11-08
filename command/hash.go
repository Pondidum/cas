package command

import (
	"bufio"
	"cas/tracing"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"sort"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

type HashCommand struct {
	Meta

	fs        fs.FS
	testInput io.ReadCloser

	algorithm string
}

func (c *HashCommand) Name() string {
	return "hash"
}

func (c *HashCommand) Synopsis() string {
	return "Generate a hash of all files passed in"
}

func (c *HashCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)

	flags.StringVar(&c.algorithm, "algorithm", "sha256", "change the hashing algorithm used")

	return flags
}

func (c *HashCommand) RunContext(ctx context.Context, args []string) error {

	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	input, err := c.selectInputSource(ctx, args)
	if err != nil {
		return tracing.Error(span, err)
	}
	defer input.Close()

	hashes, err := c.hashFiles(ctx, input)
	if err != nil {
		return tracing.Error(span, err)
	}

	sort.Strings(hashes)

	hasher, err := c.newHasher(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	for _, h := range hashes {
		if _, err := hasher.Write([]byte(h)); err != nil {
			return tracing.Error(span, err)
		}
	}

	hash := hasher.Sum(nil)

	c.Ui.Output(fmt.Sprintf("%x", hash))
	return nil
}

func (c *HashCommand) selectInputSource(ctx context.Context, args []string) (io.ReadCloser, error) {
	ctx, span := c.tr.Start(ctx, "select_input_source")
	defer span.End()

	if c.testInput != nil {
		span.SetAttributes(attribute.String("input_source", "test_input"))
		return c.testInput, nil
	}

	if len(args) > 0 {
		inputFilePath := args[0]

		span.SetAttributes(
			attribute.String("input_source", "file"),
			attribute.String("input_file", inputFilePath),
		)

		input, err := c.fs.Open(inputFilePath)
		if err != nil {
			return nil, tracing.Error(span, err)
		}

		return input, nil
	}

	span.SetAttributes(attribute.String("input_source", "stdin"))

	return os.Stdin, nil
}

var hashAlgorithms = map[string]func() hash.Hash{
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
	"md5":    md5.New,
}

func (c *HashCommand) newHasher(ctx context.Context) (hash.Hash, error) {
	ctx, span := c.tr.Start(ctx, "create_hasher")
	defer span.End()

	span.SetAttributes(attribute.String("hash_type", c.algorithm))

	createHasher, found := hashAlgorithms[c.algorithm]
	if !found {
		return nil, tracing.Error(span, fmt.Errorf("%s is not supported, try one of sha1, sha256, sha512, md5", c.algorithm))
	}

	return createHasher(), nil
}

func (c *HashCommand) hashFiles(ctx context.Context, input io.Reader) ([]string, error) {
	ctx, span := c.tr.Start(ctx, "hash_files")
	defer span.End()

	hashes := []string{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {

		filepath := scanner.Text()

		hash, err := c.hashFile(ctx, filepath)
		if err != nil {
			return nil, tracing.Error(span, err)
		}

		hashes = append(hashes, fmt.Sprintf("%s  %s\n", hash, filepath))
	}

	span.SetAttributes(attribute.Int("files_hashed", len(hashes)))

	return hashes, nil
}

func (c *HashCommand) hashFile(ctx context.Context, filepath string) (string, error) {
	ctx, span := c.tr.Start(ctx, "hash_file")
	defer span.End()

	span.SetAttributes(attribute.String("filepath", filepath))

	hasher, err := c.newHasher(ctx)
	if err != nil {
		return "", tracing.Error(span, err)
	}

	file, err := c.fs.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	span.SetAttributes(attribute.String("hash", hash))

	return hash, nil
}
