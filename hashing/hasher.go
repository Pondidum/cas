package hashing

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
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const AlgorithmSha1 = "sha1"
const AlgorithmSha256 = "sha256"
const AlgorithmSha512 = "sha512"
const AlgorithmMd5 = "md5"

var hashAlgorithms = map[string]func() hash.Hash{
	AlgorithmSha1:   sha1.New,
	AlgorithmSha256: sha256.New,
	AlgorithmSha512: sha512.New,
	AlgorithmMd5:    md5.New,
}

func AllAlgorithms() []string {
	return []string{
		AlgorithmSha1,
		AlgorithmSha256,
		AlgorithmSha512,
		AlgorithmMd5,
	}
}

func NewHasher(algorithm string) (*Hasher, error) {
	createHasher, found := hashAlgorithms[algorithm]
	if !found {
		return nil, fmt.Errorf("%s is not supported. supported algorithms: %s", algorithm, strings.Join(AllAlgorithms(), ", "))
	}

	return &Hasher{
		tr:        otel.Tracer("hasher"),
		newHasher: createHasher,
	}, nil
}

type Hasher struct {
	tr        trace.Tracer
	newHasher func() hash.Hash
}

type FileHash struct {
	Path string
	Hash string
}

func (h *Hasher) Hash(ctx context.Context, input io.Reader) (string, []FileHash, error) {
	ctx, span := h.tr.Start(ctx, "hash_input")
	defer span.End()

	hashes, err := h.hashFiles(ctx, input)
	if err != nil {
		return "", nil, tracing.Error(span, err)
	}

	hasher := h.newHasher()
	for _, f := range hashes {
		if _, err := hasher.Write([]byte(f.Hash)); err != nil {
			return "", nil, tracing.Error(span, err)
		}
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	span.SetAttributes(attribute.String("hash", hash))

	return hash, hashes, nil
}

func (h *Hasher) hashFiles(ctx context.Context, input io.Reader) ([]FileHash, error) {
	ctx, span := h.tr.Start(ctx, "hash_files")
	defer span.End()

	hashes := []FileHash{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {

		hash, err := h.hashFile(ctx, scanner.Text())
		if err != nil {
			span.SetAttributes(attribute.String("err_filepath", hash.Path))
			return nil, tracing.Error(span, err)
		}

		hashes = append(hashes, hash)
	}

	span.SetAttributes(attribute.Int("files_hashed", len(hashes)))

	return hashes, nil
}

func (h *Hasher) hashFile(ctx context.Context, filepath string) (FileHash, error) {

	fh := FileHash{
		Path: filepath,
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fh, err
	}
	defer file.Close()

	hasher := h.newHasher()

	if _, err := io.Copy(hasher, file); err != nil {
		return fh, err
	}

	fh.Hash = fmt.Sprintf("%x", hasher.Sum(nil))

	return fh, nil
}
