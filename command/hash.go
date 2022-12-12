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
	"strings"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

var DebugModeOff = "off"
var DebugModeLocal = "local"
var DebugModeStore = "store"

type HashCommand struct {
	Meta

	fs        fs.FS
	testInput io.ReadCloser

	algorithm string
	debugMode string
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
	flags.StringVar(&c.debugMode, "debug", "off", "store intermediate hashes for debugging: "+strings.Join([]string{DebugModeOff, DebugModeLocal, DebugModeStore}, " | "))

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

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	if err := c.debug(ctx, hashes, hash); err != nil {
		return tracing.Error(span, err)
	}

	c.Ui.Output(hash)
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

	// dirFS will error if a path starts with ./
	filepath = strings.TrimPrefix(filepath, "./")

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

func (c *HashCommand) debug(ctx context.Context, fileHashes []string, outputHash string) error {
	ctx, span := c.tr.Start(ctx, "debug")
	defer span.End()

	switch c.debugMode {
	case DebugModeOff:
		return nil

	case DebugModeLocal:
		return c.debugLocal(ctx, fileHashes, outputHash)

	case DebugModeStore:
		return c.debugStore(ctx, fileHashes, outputHash)

	default:
		return tracing.Errorf(span, "invalid debug mode: %s", c.debugMode)

	}
}

func (c *HashCommand) debugLocal(ctx context.Context, fileHashes []string, outputHash string) error {
	ctx, span := c.tr.Start(ctx, "debug_local")
	defer span.End()

	f, err := os.Create(fmt.Sprintf("cas-debug-%s.%s", outputHash, c.algorithm))
	if err != nil {
		return tracing.Error(span, err)
	}
	defer f.Close()

	for _, hash := range fileHashes {
		if _, err := f.WriteString(hash); err != nil {
			return tracing.Error(span, err)
		}
	}

	return nil
}

func (c *HashCommand) debugStore(ctx context.Context, fileHashes []string, outputHash string) error {
	ctx, span := c.tr.Start(ctx, "debug_local")
	defer span.End()

	f, err := os.CreateTemp("", "cas-*")
	if err != nil {
		return tracing.Error(span, err)
	}
	defer f.Close()

	for _, hash := range fileHashes {
		if _, err := f.WriteString(hash); err != nil {
			return tracing.Error(span, err)
		}
	}

	backend, err := c.createBackend(ctx)
	if err != nil {
		return tracing.Error(span, err)
	}

	storage := c.createStorage(ctx)

	if _, err := backend.StoreArtifacts(ctx, storage, outputHash, []string{f.Name()}); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}
