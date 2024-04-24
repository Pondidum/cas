package command

import (
	"cas/backends"
	"cas/backends/cache"
	"cas/backends/s3"
	"context"
	"fmt"
	"strings"
)

func parseKeyValuePairs(tags []string) (map[string]string, error) {

	m := map[string]string{}

	for _, pair := range tags {
		index := strings.Index(pair, "=")
		if index == -1 {
			return nil, fmt.Errorf("must be in the format key=value")
		}

		key := strings.TrimSpace(pair[:index])
		val := strings.TrimSpace(pair[index+1:])

		if key == "" {
			return nil, fmt.Errorf("no key specified (must be in the format key=value)")
		}
		if val == "" {
			return nil, fmt.Errorf("no value specified (must be in the format key=value)")
		}

		m[key] = val
	}

	return m, nil
}

func createBackend(ctx context.Context, backendType string) (backends.Backend, error) {
	switch strings.ToLower(backendType) {
	case "s3":
		cfg := s3.ConfigFromEnvironment()
		return cache.NewCachedBackend(s3.NewS3Backend(cfg)), nil
	}

	return nil, fmt.Errorf("unsupported backend '%s'", backendType)
}
