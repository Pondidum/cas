package s3

import (
	"cas/backends"
	"cas/localstorage"
	"cas/tracing"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tr = otel.Tracer("s3_backend")

type S3Backend struct {
	cfg    S3Config
	client *s3.Client
}

func NewS3Backend(ctx context.Context, cfg S3Config) (*S3Backend, error) {

	client, err := createClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &S3Backend{
		cfg:    cfg,
		client: client,
	}, nil
}

func createClient(ctx context.Context, cas S3Config) (*s3.Client, error) {

	opts := []func(*config.LoadOptions) error{}
	if cas.AccessKey != "" && cas.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cas.AccessKey, cas.SecretKey, "")))
	}

	if cas.Endpoint != "" {
		opts = append(opts, config.WithBaseEndpoint(cas.Endpoint))
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func EnsureBucket(ctx context.Context, cfg S3Config) error {
	client, err := createClient(ctx, cfg)
	if err != nil {
		return err
	}

	_, err = client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &cfg.BucketName,
	})

	return err
}

func (s *S3Backend) WriteMetadata(ctx context.Context, hash string, key string, value io.ReadSeeker) error {
	ctx, span := tr.Start(ctx, "write_metadata")
	defer span.End()

	span.SetAttributes(
		attribute.String("key", key),
	)

	req := &s3.PutObjectInput{
		Bucket: &s.cfg.BucketName,
		Key:    s.metadataPath(hash, key),
		Body:   value,
	}

	if _, err := s.client.PutObject(ctx, req); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}

func (s *S3Backend) ReadMetadata(ctx context.Context, hash string, keys []string) (map[string]string, error) {
	ctx, span := tr.Start(ctx, "read_metadata")
	defer span.End()

	// if no keys are passed in, we return all keys and values
	if len(keys) == 0 {
		var err error
		keys, err = s.listMetadataKeys(ctx, hash)
		if err != nil {
			return nil, tracing.Error(span, err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(keys))

	errChan := make(chan error, len(keys))
	pairs := make(map[string]string, len(keys))

	for _, k := range keys {
		go func(c context.Context, key string) {
			ctx, span := tr.Start(c, "read_"+key)
			defer span.End()
			defer wg.Done()

			span.SetAttributes(attribute.String("key", key))

			res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: &s.cfg.BucketName,
				Key:    s.metadataPath(hash, key),
			})
			if err != nil {
				// if the key doesn't exist, that isn't an error for us, just no results.
				var nokey *types.NoSuchKey
				if errors.As(err, &nokey) {
					return
				}

				errChan <- tracing.Error(span, err)
				return
			}
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			pairs[key] = string(b)

		}(ctx, k)
	}

	wg.Wait()

	if err := collectErrors(errChan); err != nil {
		return nil, tracing.Error(span, err)
	}

	return pairs, nil
}

func (s *S3Backend) listMetadataKeys(ctx context.Context, hash string) ([]string, error) {
	ctx, span := tr.Start(ctx, "list_metadata_keys")
	defer span.End()

	res, err := s.client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: &s.cfg.BucketName,
		Prefix: s.metadataPath(hash, ""),
	})

	if err != nil {
		return nil, tracing.Error(span, err)
	}

	keys := make([]string, len(res.Contents))

	for i, o := range res.Contents {
		keys[i] = path.Base(*o.Key)
	}

	return keys, nil
}

func (s *S3Backend) hasMetadata(ctx context.Context, hash string, key string) (bool, error) {
	ctx, span := tr.Start(ctx, "has_metadata")
	defer span.End()

	span.SetAttributes(
		attribute.String("hash", hash),
		attribute.String("key", key),
	)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.cfg.BucketName,
		Key:    s.metadataPath(hash, key),
	})
	if err != nil {
		if _, ok := err.(*smithy.OperationError); ok {
			return false, nil
		}

		return false, tracing.Error(span, err)
	}

	return true, nil
}

func (s *S3Backend) StoreArtifacts(ctx context.Context, hash string, files []*localstorage.LocalFile) ([]string, error) {

	ctx, span := tr.Start(ctx, "store_artifacts")
	defer span.End()

	_, found, err := backends.ReadTimestamp(ctx, s, hash)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	span.SetAttributes(attribute.Bool("has_timestamp", found))

	if !found {
		if err := backends.CreateHash(ctx, s, hash, time.Now()); err != nil {
			return nil, tracing.Error(span, err)
		}

		span.SetAttributes(attribute.Bool("hash_created", true))
	}

	wg := sync.WaitGroup{}
	wg.Add(len(files))

	errChan := make(chan error, len(files))
	writtenChan := make(chan string, len(files))

	for _, lf := range files {

		go func(c context.Context, localFile *localstorage.LocalFile) {
			ctx, span := tr.Start(c, "store_"+path.Base(localFile.Path))
			defer span.End()
			defer localFile.Close()
			defer wg.Done()

			s3path := s.artifactPath(hash, localFile.Path)
			span.SetAttributes(
				attribute.String("local_path", localFile.Path),
				attribute.String("remote_path", s3path),
			)

			sha, err := hashFile(ctx, localFile.Content)
			if err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			span.SetAttributes(attribute.String("local_hash", sha))

			if _, err := localFile.Content.Seek(0, 0); err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			req := &s3.PutObjectInput{
				Bucket: &s.cfg.BucketName,
				Key:    &s3path,
				Body:   localFile.Content,
				Metadata: map[string]string{
					"sha1": sha,
				},
			}

			if _, err := s.client.PutObject(ctx, req); err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			writtenChan <- localFile.Path
		}(ctx, lf)
	}

	wg.Wait()

	written := collectArray(writtenChan)

	if err := collectErrors(errChan); err != nil {
		return written, tracing.Error(span, err)
	}

	return written, nil
}

func hashFile(ctx context.Context, file io.Reader) (string, error) {

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (s *S3Backend) FetchArtifacts(ctx context.Context, hash string) ([]*backends.RemoteFile, error) {
	return nil, fmt.Errorf("not implemented, you should use the cachebackend wrapper")
}

func (s *S3Backend) FetchArtifact(ctx context.Context, hash string, name string) (*backends.RemoteFile, error) {
	ctx, span := tr.Start(ctx, "fetch_artifact")
	defer span.End()

	span.SetAttributes(attribute.String("artifact_name", name))

	remotePath := s.artifactPath(hash, name)

	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.cfg.BucketName,
		Key:    &remotePath,
	})
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	ts, _, err := backends.ReadTimestamp(ctx, s, hash)
	if err != nil {
		return nil, tracing.Error(span, err)
	}

	return &backends.RemoteFile{
		Name:      name,
		Content:   res.Body,
		Timestamp: ts,
	}, nil

}

func (s *S3Backend) ListArtifacts(ctx context.Context, hash string) ([]string, error) {
	ctx, span := tr.Start(ctx, "list_artifact_keys")
	defer span.End()

	artifactPath := s.artifactPath(hash, "")

	res, err := s.client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: &s.cfg.BucketName,
		Prefix: &artifactPath,
	})

	if err != nil {
		return nil, tracing.Error(span, err)
	}

	keys := make([]string, len(res.Contents))

	for i, o := range res.Contents {

		name, err := filepath.Rel(artifactPath, *o.Key)
		if err != nil {
			return nil, tracing.Error(span, err)
		}

		keys[i] = name
	}

	return keys, nil
}

func (s *S3Backend) metadataPath(hash string, key string) *string {
	p := path.Join(s.cfg.PathPrefix, "meta", hash, key)
	return &p
}

func (s *S3Backend) artifactPath(hash string, artifactPath string) string {
	return path.Join(s.cfg.PathPrefix, "artifact", hash, artifactPath)
}
