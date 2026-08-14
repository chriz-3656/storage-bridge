package local

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/storage-bridge/core/pkg/storage"
)

type Provider struct {
	Root string
}

func New(root string) (*Provider, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Provider{Root: abs}, nil
}

func (p *Provider) Name() string {
	return "local"
}

func (p *Provider) resolve(path string) (string, error) {
	fullPath := filepath.Join(p.Root, filepath.FromSlash(path))
	return fullPath, nil
}

func (p *Provider) Stat(ctx context.Context, path string) (*storage.Entry, error) {
	fullPath, err := p.resolve(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		if os.IsPermission(err) {
			return nil, storage.ErrPermissionDenied
		}
		return nil, err
	}

	return &storage.Entry{
		Path:    path,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

type localIterator struct {
	entries []fs.DirEntry
	idx     int
	dirPath string
}

func (l *localIterator) Next(ctx context.Context) (*storage.Entry, error) {
	if l.idx >= len(l.entries) {
		return nil, io.EOF
	}
	
	de := l.entries[l.idx]
	l.idx++
	
	info, err := de.Info()
	if err != nil {
		// skip errors for individual files in listing (like permission denied)
		return &storage.Entry{Path: filepath.ToSlash(filepath.Join(l.dirPath, de.Name())), IsDir: de.IsDir()}, nil
	}

	return &storage.Entry{
		Path:    filepath.ToSlash(filepath.Join(l.dirPath, de.Name())),
		IsDir:   de.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

func (p *Provider) List(ctx context.Context, path string) (storage.Iterator, error) {
	fullPath, err := p.resolve(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return &localIterator{
		entries: entries,
		dirPath: path,
	}, nil
}

func (p *Provider) Get(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error) {
	fullPath, err := p.resolve(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			f.Close()
			return nil, err
		}
	}

	if length < 0 {
		return f, nil
	}

	return struct {
		io.Reader
		io.Closer
	}{
		io.LimitReader(f, length),
		f,
	}, nil
}

func (p *Provider) Put(ctx context.Context, path string, in io.Reader, size int64, modTime time.Time) error {
	fullPath, err := p.resolve(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, in); err != nil {
		return err
	}

	if !modTime.IsZero() {
		os.Chtimes(fullPath, modTime, modTime)
	}

	return nil
}

func (p *Provider) Remove(ctx context.Context, path string) error {
	fullPath, err := p.resolve(path)
	if err != nil {
		return err
	}

	err = os.RemoveAll(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}

func (p *Provider) SpaceUsed(ctx context.Context, path string) (int64, error) {
	fullPath, err := p.resolve(path)
	if err != nil {
		return 0, err
	}

	var total int64
	err = filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})

	return total, err
}
