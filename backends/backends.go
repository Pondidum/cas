package backends

import "context"

type Backend interface {
	WriteMetadata(ctx context.Context, hash string, data map[string]string) (map[string]string, error)
}
