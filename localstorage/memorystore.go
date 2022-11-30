package localstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
)

type closableBuffer struct {
	*bytes.Reader
}

func (b *closableBuffer) Close() error {
	return nil
}

type MemoryStorage struct {
	Store map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Store: map[string][]byte{},
	}
}

func (m *MemoryStorage) ReadFile(ctx context.Context, p string) (io.ReadCloser, error) {
	if content, found := m.Store[p]; found {
		return &closableBuffer{Reader: bytes.NewReader(content)}, nil
	}

	return nil, fmt.Errorf("file not found: %s", p)
}

func (m *MemoryStorage) WriteFile(ctx context.Context, path string, content io.Reader) error {
	b, err := ioutil.ReadAll(content)
	if err != nil {
		return err
	}

	m.Store[path] = b

	return nil
}
