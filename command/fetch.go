package command

import (
	"bufio"
	"bytes"
	"cas/backends"
	"cas/tracing"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

func NewFetchCommand(ui cli.Ui) *FetchCommand {
	cmd := &FetchCommand{}
	cmd.Meta = NewMeta(ui, cmd)
	return cmd
}

type FetchCommand struct {
	Meta

	algorithm string
	statePath string

	// for testing hashing on streams of data
	testInput io.ReadCloser

	// for skipping hashing to use a specific value
	testHash string
}

func (c *FetchCommand) Name() string {
	return "Fetch"
}

func (c *FetchCommand) Synopsis() string {
	return "Fetches state and artifacts for a set of files"
}

func (c *FetchCommand) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(c.Name(), pflag.ContinueOnError)

	flags.StringVar(&c.statePath, "state-path", ".cas/state", "the directory to hold local state")
	flags.StringVar(&c.algorithm, "algorithm", "sha256", "change the hashing algorithm used")

	return flags
}

func (c *FetchCommand) RunContext(ctx context.Context, args []string) error {
	ctx, span := c.tr.Start(ctx, "run")
	defer span.End()

	hash, err := c.hashInput(ctx, args)
	if err != nil {
		return tracing.Error(span, err)
	}

	backend, err := c.createBackend(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	ts, err := c.readTimestamp(ctx, backend, hash)
	if err != nil {
		return tracing.Error(span, err)
	}

	isExistingHash := ts != nil
	span.SetAttributes(attribute.Bool("existing_hash", isExistingHash))

	if ts == nil {
		now := time.Now()
		ts = &now
	}

	storage := c.createStorage(ctx)
	statePath := path.Join(c.statePath, hash)

	if err := storage.WriteFile(ctx, statePath, *ts, &bytes.Buffer{}); err != nil {
		return tracing.Error(span, err)
	}

	if isExistingHash {
		writeFile := func(ctx context.Context, relPath string, content io.Reader) error {
			return storage.WriteFile(ctx, relPath, *ts, content)
		}

		if err := backend.FetchArtifacts(ctx, hash, writeFile); err != nil {
			return tracing.Error(span, err)
		}

	} else {
		_, err := backend.WriteMetadata(ctx, hash, map[string]string{"@timestamp": fmt.Sprintf("%v", ts.Unix())})
		if err != nil {
			return tracing.Error(span, err)
		}
	}

	c.Ui.Output(statePath)

	return nil
}

func (c *FetchCommand) readTimestamp(ctx context.Context, backend backends.Backend, hash string) (*time.Time, error) {
	ctx, span := c.tr.Start(ctx, "read_timestamp")
	defer span.End()

	meta, err := backend.ReadMetadata(ctx, hash, []string{"@timestamp"})
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	timestampAnnotation, found := meta["@timestamp"]
	if !found {
		return nil, nil
	}

	seconds, err := strconv.ParseInt(timestampAnnotation, 10, 64)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	ts := time.Unix(seconds, 0)

	return &ts, nil
}

func (c *FetchCommand) selectInputSource(ctx context.Context, args []string) (io.ReadCloser, error) {
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

		input, err := os.Open(inputFilePath)
		if err != nil {
			return nil, tracing.Error(span, err)
		}

		return input, nil
	}

	span.SetAttributes(attribute.String("input_source", "stdin"))

	return os.Stdin, nil
}

func (c *FetchCommand) hashInput(ctx context.Context, args []string) (string, error) {
	ctx, span := c.tr.Start(ctx, "hash_input")
	defer span.End()

	if c.testHash != "" {
		return c.testHash, nil
	}

	input, err := c.selectInputSource(ctx, args)
	if err != nil {
		return "", tracing.Error(span, err)
	}
	defer input.Close()

	span.SetAttributes(attribute.String("hash_type", c.algorithm))

	hasher, err := c.newHasher()
	if err != nil {
		return "", tracing.Error(span, err)
	}

	hashes, err := c.hashFiles(ctx, input)
	if err != nil {
		return "", tracing.Error(span, err)
	}

	for _, h := range hashes {
		if _, err := hasher.Write([]byte(h)); err != nil {
			return "", tracing.Error(span, err)
		}
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	span.SetAttributes(attribute.String("hash", hash))

	return hash, nil
}

var hashAlgorithms = map[string]func() hash.Hash{
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
	"md5":    md5.New,
}

func (c *FetchCommand) newHasher() (hash.Hash, error) {
	createHasher, found := hashAlgorithms[c.algorithm]
	if !found {
		return nil, fmt.Errorf("%s is not supported, try one of sha1, sha256, sha512, md5", c.algorithm)
	}

	return createHasher(), nil
}

func (c *FetchCommand) hashFiles(ctx context.Context, input io.Reader) ([]string, error) {
	ctx, span := c.tr.Start(ctx, "hash_files")
	defer span.End()

	hashes := []string{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {

		filepath := scanner.Text()

		hash, err := c.hashFile(ctx, filepath)
		if err != nil {
			span.SetAttributes(attribute.String("err_filepath", filepath))
			return nil, tracing.Error(span, err)
		}

		hashes = append(hashes, fmt.Sprintf("%s  %s\n", hash, filepath))
	}

	span.SetAttributes(attribute.Int("files_hashed", len(hashes)))

	return hashes, nil
}

func (c *FetchCommand) hashFile(ctx context.Context, filepath string) (string, error) {

	hasher, err := c.newHasher()
	if err != nil {
		return "", err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
