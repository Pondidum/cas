package s3

import (
	"cas/tracing"
	"context"
	"fmt"
	"io"
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

type S3Backend struct {
	cfg    S3Config
	client *s3.Client
}

func NewS3Backend(cfg S3Config) *S3Backend {
	return &S3Backend{
		cfg:    cfg,
		client: createClient(cfg),
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
	written := make(map[string]string, len(data))

	for key, value := range data {

		go func(c context.Context, k, v string) {
			ctx, span := tr.Start(c, "write_"+k)
			defer span.End()

			span.SetAttributes(
				attribute.String("key", k),
				attribute.String("value", v),
			)

			req := &s3.PutObjectInput{
				Bucket: &s.cfg.BucketName,
				Key:    s.metadataPath(hash, k),
				Body:   strings.NewReader(v),
			}

			if _, err := s.client.PutObject(ctx, req); err != nil {
				errChan <- tracing.Error(span, err)
			} else {
				written[k] = v
			}

			wg.Done()
		}(ctx, key, value)
	}

	wg.Wait()

	if err := collectErrors(errChan); err != nil {
		return written, tracing.Error(span, err)
	}

	return written, nil
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

func (s *S3Backend) metadataPath(hash string, key string) *string {
	p := path.Join(s.cfg.PathPrefix, "meta", hash, key)
	return &p
}

func (s *S3Backend) StoreArtifact(ctx context.Context, hash string, path string, content io.Reader) error {
	ctx, span := tr.Start(ctx, "store_artifact")
	defer span.End()

	// force the timestamp to exist
	if _, err := s.WriteMetadata(ctx, hash, map[string]string{}); err != nil {
		return tracing.Error(span, err)
	}

	req := &s3.PutObjectInput{
		Bucket:      &s.cfg.BucketName,
		Key:         s.artifactPath(hash, path),
		Body:        content,
		ContentType: &binaryMimeType,
	}

	if _, err := s.client.PutObject(ctx, req); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}

func (s *S3Backend) artifactPath(hash string, artifactPath string) *string {
	p := path.Join(s.cfg.PathPrefix, "artifact", hash, artifactPath)
	return &p
}
