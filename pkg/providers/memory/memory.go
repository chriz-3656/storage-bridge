package memory

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/storage-bridge/core/pkg/storage"
)

type Provider struct {
	mu    sync.RWMutex
	files map[string]*memFile
}

type memFile struct {
	data    []byte
	modTime time.Time
	isDir   bool
}

func New() *Provider {
	return &Provider{
		files: make(map[string]*memFile),
	}
}

func (p *Provider) Name() string {
	return "memory"
}

func (p *Provider) Stat(ctx context.Context, path string) (*storage.Entry, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	path = strings.TrimPrefix(path, "/")
	
	if path == "" {
		return &storage.Entry{Path: "", IsDir: true, ModTime: time.Now()}, nil
	}

	f, ok := p.files[path]
	if !ok {
		// check if it's a directory
		prefix := path + "/"
		for k := range p.files {
			if strings.HasPrefix(k, prefix) {
				return &storage.Entry{Path: path, IsDir: true, ModTime: time.Now()}, nil
			}
		}
		return nil, storage.ErrNotFound
	}

	return &storage.Entry{
		Path:    path,
		IsDir:   f.isDir,
		Size:    int64(len(f.data)),
		ModTime: f.modTime,
	}, nil
}

type memIterator struct {
	entries []*storage.Entry
	idx     int
}

func (m *memIterator) Next(ctx context.Context) (*storage.Entry, error) {
	if m.idx >= len(m.entries) {
		return nil, io.EOF
	}
	e := m.entries[m.idx]
	m.idx++
	return e, nil
}

func (p *Provider) List(ctx context.Context, path string) (storage.Iterator, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	path = strings.TrimPrefix(path, "/")
	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	dirs := make(map[string]bool)
	var entries []*storage.Entry

	for k, f := range p.files {
		if path == "" || strings.HasPrefix(k, prefix) {
			rel := strings.TrimPrefix(k, prefix)
			parts := strings.SplitN(rel, "/", 2)
			if len(parts) > 1 {
				dirName := parts[0]
				if !dirs[dirName] {
					dirs[dirName] = true
					entries = append(entries, &storage.Entry{
						Path:  prefix + dirName,
						IsDir: true,
					})
				}
			} else {
				entries = append(entries, &storage.Entry{
					Path:    k,
					IsDir:   f.isDir,
					Size:    int64(len(f.data)),
					ModTime: f.modTime,
				})
			}
		}
	}

	return &memIterator{entries: entries}, nil
}

func (p *Provider) Get(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	path = strings.TrimPrefix(path, "/")
	f, ok := p.files[path]
	if !ok {
		return nil, storage.ErrNotFound
	}
	if f.isDir {
		return nil, storage.ErrIsDirectory
	}

	var data []byte
	if offset > int64(len(f.data)) {
		offset = int64(len(f.data))
	}
	if length < 0 || offset+length > int64(len(f.data)) {
		data = f.data[offset:]
	} else {
		data = f.data[offset : offset+length]
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (p *Provider) Put(ctx context.Context, path string, in io.Reader, size int64, modTime time.Time) error {
	path = strings.TrimPrefix(path, "/")
	data, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.files[path] = &memFile{
		data:    data,
		modTime: modTime,
		isDir:   false,
	}

	return nil
}

func (p *Provider) Remove(ctx context.Context, path string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	path = strings.TrimPrefix(path, "/")
	
	if _, ok := p.files[path]; ok {
		delete(p.files, path)
		return nil
	}

	// Delete prefix for dirs
	prefix := path + "/"
	deleted := false
	for k := range p.files {
		if strings.HasPrefix(k, prefix) {
			delete(p.files, k)
			deleted = true
		}
	}

	if !deleted {
		return storage.ErrNotFound
	}
	return nil
}
