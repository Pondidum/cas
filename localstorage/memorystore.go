package localstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"
	"time"
)

type closableBuffer struct {
	*bytes.Reader
}

func (b *closableBuffer) Close() error {
	return nil
}

type MemoryStorage struct {
	Store    map[string][]byte
	Modified map[string]time.Time
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Store:    map[string][]byte{},
		Modified: map[string]time.Time{},
	}
}

func (m *MemoryStorage) ListFiles(ctx context.Context, p string) ([]string, error) {

	p = strings.ToLower(p)
	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}

	files := []string{}

	for key := range m.Store {
		if strings.HasPrefix(strings.ToLower(key), p) {
			files = append(files, key)
		}
	}

	sort.Strings(files)

	return files, nil
}

func (m *MemoryStorage) ReadFile(ctx context.Context, p string) (io.ReadCloser, error) {
	if content, found := m.Store[p]; found {
		return &closableBuffer{Reader: bytes.NewReader(content)}, nil
	}

	return nil, fmt.Errorf("file not found: %s", p)
}

func (m *MemoryStorage) WriteFile(ctx context.Context, path string, timestamp time.Time, content io.Reader) error {
	b, err := ioutil.ReadAll(content)
	if err != nil {
		return err
	}

	m.Store[path] = b
	m.Modified[path] = timestamp

	return nil
}
