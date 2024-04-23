package backends

import (
	"cas/localstorage"
	"context"
	"io"
	"time"
)

type Backend interface {
	WriteMetadata(ctx context.Context, hash string, key string, value io.ReadSeeker) error
	ReadMetadata(ctx context.Context, hash string, keys []string) (map[string]string, error)

	StoreArtifacts(ctx context.Context, hash string, files []*localstorage.LocalFile) ([]string, error)

	ListArtifacts(ctx context.Context, hash string) ([]string, error)
	FetchArtifact(ctx context.Context, hash string, name string) (*RemoteFile, error)
	FetchArtifacts(ctx context.Context, hash string) ([]*RemoteFile, error)
}

type RemoteFile struct {
	Name      string
	Timestamp time.Time
	Content   io.ReadCloser
}

func (rf *RemoteFile) Close() error {
	return rf.Content.Close()
}
