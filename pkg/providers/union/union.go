package union

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/storage-bridge/core/pkg/storage"
)

type Policy string

const (
	PolicyFirst Policy = "first" // Write to the first upstream
	PolicyAll   Policy = "all"   // Mirror writes to all upstreams
)

// Provider implements the storage.Provider interface by delegating
// to a list of upstream providers.
type Provider struct {
	Upstreams []storage.Provider
	Policy    Policy
}

// New creates a new union provider.
func New(upstreams []storage.Provider, policy Policy) *Provider {
	if policy == "" {
		policy = PolicyFirst
	}
	return &Provider{
		Upstreams: upstreams,
		Policy:    policy,
	}
}

func (p *Provider) Stat(ctx context.Context, path string) (storage.FileEntry, error) {
	for _, up := range p.Upstreams {
		entry, err := up.Stat(ctx, path)
		if err == nil {
			return entry, nil
		}
	}
	return storage.FileEntry{}, fmt.Errorf("file not found in any upstream: %s", path)
}

// unionIterator merges results from multiple iterators
type unionIterator struct {
	ctx      context.Context
	iters    []storage.Iterator
	currIdx  int
	seen     map[string]bool
}

func (u *unionIterator) Next(ctx context.Context) (storage.FileEntry, error) {
	for u.currIdx < len(u.iters) {
		entry, err := u.iters[u.currIdx].Next(ctx)
		if err == io.EOF {
			u.currIdx++
			continue
		}
		if err != nil {
			return storage.FileEntry{}, err
		}
		
		// Deduplicate
		if u.seen[entry.Path] {
			continue
		}
		u.seen[entry.Path] = true
		return entry, nil
	}
	return storage.FileEntry{}, io.EOF
}

func (p *Provider) List(ctx context.Context, prefix string) (storage.Iterator, error) {
	var iters []storage.Iterator
	for _, up := range p.Upstreams {
		iter, err := up.List(ctx, prefix)
		if err == nil && iter != nil {
			iters = append(iters, iter)
		}
	}
	
	if len(iters) == 0 {
		return nil, fmt.Errorf("failed to list any upstreams")
	}
	
	return &unionIterator{
		ctx:   ctx,
		iters: iters,
		seen:  make(map[string]bool),
	}, nil
}

func (p *Provider) Get(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error) {
	for _, up := range p.Upstreams {
		rc, err := up.Get(ctx, path, offset, length)
		if err == nil {
			return rc, nil
		}
	}
	return nil, fmt.Errorf("could not get file from any upstream: %s", path)
}

func (p *Provider) Put(ctx context.Context, path string, reader io.Reader, size int64, modTime time.Time) error {
	if len(p.Upstreams) == 0 {
		return fmt.Errorf("no upstreams configured")
	}
	
	if p.Policy == PolicyFirst {
		// Just write to the first one that succeeds
		for _, up := range p.Upstreams {
			// Note: if reader is consumed and fails, subsequent attempts will fail unless we seek/buffer.
			// For simplicity in MVP, we just try the first and fail if it fails.
			err := up.Put(ctx, path, reader, size, modTime)
			if err == nil {
				return nil
			}
			return err // Return immediate error because reader is consumed
		}
	}
	
	return fmt.Errorf("unsupported policy or all upstreams failed")
}

func (p *Provider) Remove(ctx context.Context, path string) error {
	var lastErr error
	removedAny := false
	
	for _, up := range p.Upstreams {
		err := up.Remove(ctx, path)
		if err == nil {
			removedAny = true
		} else {
			lastErr = err
		}
	}
	
	if !removedAny && lastErr != nil {
		return lastErr
	}
	return nil
}

func (p *Provider) SpaceUsed(ctx context.Context, path string) (int64, error) {
	var total int64
	for _, up := range p.Upstreams {
		used, err := up.SpaceUsed(ctx, path)
		if err == nil {
			total += used
		}
	}
	return total, nil
}

func (p *Provider) Mkdir(ctx context.Context, path string) error {
	// Create directory in all upstreams to ensure consistency
	var lastErr error
	createdAny := false
	
	for _, up := range p.Upstreams {
		err := up.Mkdir(ctx, path)
		if err == nil {
			createdAny = true
		} else {
			lastErr = err
		}
	}
	
	if !createdAny && lastErr != nil {
		return lastErr
	}
	return nil
}

func (p *Provider) Move(ctx context.Context, src string, dest string) error {
	// Move it wherever it exists
	var lastErr error
	movedAny := false
	
	for _, up := range p.Upstreams {
		// Only attempt move if it exists here
		_, err := up.Stat(ctx, src)
		if err == nil {
			err = up.Move(ctx, src, dest)
			if err == nil {
				movedAny = true
			} else {
				lastErr = err
			}
		}
	}
	
	if !movedAny {
		if lastErr != nil {
			return lastErr
		}
		return fmt.Errorf("source file not found in any upstream to move")
	}
	return nil
}
