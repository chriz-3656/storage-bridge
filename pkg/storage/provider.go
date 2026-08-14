package storage

import (
	"context"
	"io"
	"time"
)

// Provider represents a configured storage backend.
type Provider interface {
	Name() string
	Stat(ctx context.Context, path string) (*Entry, error)
	List(ctx context.Context, path string) (Iterator, error)
	Get(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error)
	Put(ctx context.Context, path string, in io.Reader, size int64, modTime time.Time) error
	Remove(ctx context.Context, path string) error
}

type Entry struct {
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

type Iterator interface {
	Next(ctx context.Context) (*Entry, error)
}
