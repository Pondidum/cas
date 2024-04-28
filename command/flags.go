package command

import (
	"cas/backends"
	"cas/backends/cache"
	"cas/backends/s3"
	"cas/config"
	"context"
	"fmt"
	"strings"
)

const BackendEnvVar = "CAS_BACKEND"

func NewBackendConfiguration() *BackendConfiguration {
	return &BackendConfiguration{
		s3: s3.S3Config{},
	}
}

type BackendConfiguration struct {
	name string

	s3 s3.S3Config
}

func (bc *BackendConfiguration) Flags() []*config.ConfigGroup {

	own := config.NewConfigGroup("backend")
	own.StringFlag(&bc.name, "backend", BackendEnvVar, "s3", "the backend to use for artifacts")

	return []*config.ConfigGroup{
		own,
		bc.s3.Flags(),
		// other backend flag sets here
	}
}

func (bc *BackendConfiguration) Create(ctx context.Context) (backends.Backend, error) {
	switch strings.ToLower(bc.name) {
	case "s3":
		return cache.NewCachedBackend(s3.NewS3Backend(bc.s3)), nil
	}

	return nil, fmt.Errorf("unsupported backend '%s'", bc.name)
}

func globalFlags() *config.ConfigGroup {
	flags := config.NewConfigGroup("global")

	// flags.StringVar(&c.statePath, "state-path", ".cas/state", "the directory to hold local state")
	format := ""
	flags.StringFlag(&format, "format", "CAS_FORMAT", "text", "which format to write to stdout")

	return flags

}
