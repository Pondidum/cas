package s3

import (
	"cas/backends"
	"cas/tracing"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const MetadataTimeStamp = "@timestamp"

var tr = otel.Tracer("s3_backend")
var binaryMimeType = "application/octet-stream"

type ReadFunc func(ctx context.Context, p string) (io.ReadCloser, error)
type WriteFunc func(ctx context.Context, path string, content io.Reader) (string, error)

type S3Backend struct {
	cfg    S3Config
	client *s3.Client

	local backends.Storage
}

func NewS3Backend(cfg S3Config, storage backends.Storage) *S3Backend {
	return &S3Backend{
		cfg:    cfg,
		client: createClient(cfg),
		local:  storage,
	}
}

func createClient(cfg S3Config) *s3.Client {
	opts := []func(*s3.Options){}

	if cfg.Endpoint != "" {
		opts = append(opts, s3.WithEndpointResolver(s3.EndpointResolverFunc(func(region string, options s3.EndpointResolverOptions) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       cfg.Region,
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		})))
	}

	return s3.New(s3.Options{
		Region: cfg.Region,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.AccessKey,
				SecretAccessKey: cfg.SecretKey,
			}, nil
		}),
	}, opts...)
}

func EnsureBucket(ctx context.Context, cfg S3Config) error {
	client := createClient(cfg)
	_, err := client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &cfg.BucketName,
	})

	return err
}

func (s *S3Backend) WriteMetadata(ctx context.Context, hash string, data map[string]string) (map[string]string, error) {
	ctx, span := tr.Start(ctx, "write_metadata")
	defer span.End()

	found, err := s.hasMetadata(ctx, hash, MetadataTimeStamp)
	if err != nil {
		return nil, tracing.Error(span, err)
	}
	if !found {
		data[MetadataTimeStamp] = fmt.Sprintf("%v", time.Now().Unix())
	}

	wg := sync.WaitGroup{}
	wg.Add(len(data))

	errChan := make(chan error, len(data))
	writtenChan := make(chan pair, len(data))

	for k, v := range data {

		go func(c context.Context, key, value string) {
			ctx, span := tr.Start(c, "write_"+key)
			defer span.End()
			defer wg.Done()

			span.SetAttributes(
				attribute.String("key", key),
				attribute.String("value", value),
			)

			req := &s3.PutObjectInput{
				Bucket: &s.cfg.BucketName,
				Key:    s.metadataPath(hash, key),
				Body:   strings.NewReader(value),
			}

			if _, err := s.client.PutObject(ctx, req); err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			writtenChan <- pair{key, value}

		}(ctx, k, v)
	}

	wg.Wait()

	written := collectMap(writtenChan)

	if err := collectErrors(errChan); err != nil {
		return written, tracing.Error(span, err)
	}

	return written, nil
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

func (s *S3Backend) StoreArtifacts(ctx context.Context, hash string, paths []string) ([]string, error) {
	ctx, span := tr.Start(ctx, "store_artifacts")
	defer span.End()

	// force the timestamp to exist
	if _, err := s.WriteMetadata(ctx, hash, map[string]string{}); err != nil {
		return nil, tracing.Error(span, err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(paths))

	errChan := make(chan error, len(paths))
	writtenChan := make(chan string, len(paths))

	for _, p := range paths {

		go func(c context.Context, filePath string) {
			ctx, span := tr.Start(c, "store_"+path.Base(filePath))
			defer span.End()
			defer wg.Done()

			s3path := s.artifactPath(hash, filePath)
			span.SetAttributes(
				attribute.String("local_path", filePath),
				attribute.String("remote_path", s3path),
			)

			content, err := s.local.ReadFile(ctx, filePath)
			if err != nil {
				errChan <- tracing.Error(span, err)
				return
			}
			defer content.Close()

			req := &s3.PutObjectInput{
				Bucket: &s.cfg.BucketName,
				Key:    &s3path,
				Body:   content,
			}

			if _, err := s.client.PutObject(ctx, req); err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			writtenChan <- filePath
		}(ctx, p)
	}

	wg.Wait()

	written := collectArray(writtenChan)

	if err := collectErrors(errChan); err != nil {
		return written, tracing.Error(span, err)
	}

	return written, nil
}

func (s *S3Backend) FetchArtifacts(ctx context.Context, hash string, paths []string) ([]string, error) {
	ctx, span := tr.Start(ctx, "fetch_artifacts")
	defer span.End()

	if len(paths) == 0 {
		var err error
		paths, err = s.listArtifactKeys(ctx, hash)
		if err != nil {
			return nil, tracing.Error(span, err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(paths))

	errChan := make(chan error, len(paths))
	writtenChan := make(chan string, len(paths))

	for _, p := range paths {
		go func(ctx context.Context, filePath string) {
			ctx, span := tr.Start(ctx, "fetch_"+path.Base(filePath))
			defer span.End()
			defer wg.Done()

			remotePath := s.artifactPath(hash, filePath)
			span.SetAttributes(attribute.String("remote_path", remotePath))

			res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: &s.cfg.BucketName,
				Key:    &remotePath,
			})
			if err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			defer res.Body.Close()
			writtenPath, err := s.local.WriteFile(ctx, filePath, res.Body)
			if err != nil {
				errChan <- tracing.Error(span, err)
				return
			}

			writtenChan <- writtenPath

		}(ctx, p)
	}

	wg.Wait()

	written := collectArray(writtenChan)

	if err := collectErrors(errChan); err != nil {
		return written, tracing.Error(span, err)
	}

	return written, nil
}

func (s *S3Backend) listArtifactKeys(ctx context.Context, hash string) ([]string, error) {
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
		keys[i] = strings.TrimPrefix(*o.Key, artifactPath)
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
