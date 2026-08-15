# Storage Abstraction

Based on the investigation of rclone's `fs.Fs` and `fs.Object`, we can design a minimalist storage abstraction tailored for a lightweight engine like `storage-bridge`.

## Rclone's Approach
Rclone uses a heavy abstraction:
- **`fs.Fs`**: `List`, `NewObject`, `Put`, `Mkdir`, `Rmdir`, `Features`, `Hashes`.
- **`fs.Object`**: `SetModTime`, `Open`, `Update`, `Remove`.
- Dozens of optional interfaces (e.g., `fs.SetTierer`, `fs.MimeTyper`).

## Do We Need Everything?
- **List, Get (Open), Put, Remove**: **YES**. These are fundamental to any storage system.
- **Stat (NewObject/Info)**: **YES**. Needed to check existence and size.
- **Copy, Move, Server-Side-Copy**: **NO** for a minimal core interface. These can be implemented as a fallback (Get + Put + Remove) by the engine if the provider doesn't support them.
- **Mkdir, Rmdir**: **NO/OPTIONAL**. Many modern object stores (S3, GCS) do not have true directories; they use prefix routing. We can abstract away directories for a minimal key-value object interface.
- **Quota, Share, Search**: **NO**. These are highly provider-specific and rarely universal.
- **Streaming**: **YES**. Uploads and downloads must support streaming (`io.Reader`/`io.ReadCloser`) to handle large files without exhausting RAM.

## Proposed Minimal Interface (storage-bridge)

The minimal interface treats storage as a simple asynchronous Key-Value blob store. Directories are just prefixes in keys.

```go
package storage

import (
	"context"
	"io"
	"time"
)

// Provider represents a connection to a specific storage backend (e.g., an S3 bucket).
type Provider interface {
	// Put streams the contents of 'in' to the specified path.
	// size can be -1 if unknown.
	Put(ctx context.Context, path string, in io.Reader, size int64) (File, error)

	// Get returns a reader for the file at the path.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Stat retrieves metadata about a file. Returns ErrNotFound if missing.
	Stat(ctx context.Context, path string) (File, error)

	// Delete removes the file at the path.
	Delete(ctx context.Context, path string) error

	// List returns a list of files matching the prefix.
	// For a minimalist approach, this could return a channel or paginated response.
	List(ctx context.Context, prefix string) ([]File, error)
}

// File represents metadata about a stored object.
type File interface {
	Path() string
	Size() int64
	ModTime() time.Time
}
```

This drastically reduces implementation complexity for new providers compared to rclone, by stripping away directory management and optional metadata setters.
