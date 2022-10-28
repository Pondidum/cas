package s3

import (
	"cas/tracing"
	"context"
	"path"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.opentelemetry.io/otel"
)

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

func (s *S3Backend) WriteMetadata(ctx context.Context, hash string, data map[string]string) error {
	ctx, span := tr.Start(ctx, "write_metadata")
	defer span.End()

	wg := sync.WaitGroup{}
	wg.Add(len(data))

	errChan := make(chan error, len(data))

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
			}

			wg.Done()
		}(ctx, key, value)
	}

	wg.Wait()

	if err := collectErrors(errChan); err != nil {
		return tracing.Error(span, err)
	}

	return nil
}

func (s *S3Backend) metadataPath(hash string, key string) *string {
	p := path.Join(s.cfg.PathPrefix, "meta", hash, key)
	return &p
}
