# Capability System

## Rclone's Capability Model
Rclone handles capabilities through a two-pronged approach:

1. **`fs.Features` Struct**: `fs.Fs` inherits `fs.Info`, which has a `Features() *fs.Features` method. This returns a struct with boolean flags for specific features and function pointers for optimizations (e.g., `Copy`, `Move`, `DirMove`, `Purge`). 
2. **Optional Interfaces**: Rclone aggressively uses Go's interface type assertion. `fs.Object` and `fs.Fs` instances are checked at runtime to see if they implement optional interfaces (e.g., `fs.MimeTyper`, `fs.SetTierer`, `fs.IDer` in `fs/types.go`).

## Proposed Capability Model for `storage-bridge`

For a lightweight engine, capabilities should be divided into **Core** (mandatory for the provider to implement) and **Optional** (enhancements that the engine will use if available, but can fall back to generic implementations if not).

### Core Capabilities (Mandatory)
Every provider must implement these via the primary `Provider` interface.
- **Upload (`Put`)**
- **Download (`Get`)**
- **Delete (`Delete`)**
- **List (`List`)**
- **Stream**: Implicitly required by `Put` and `Get` accepting/returning `io.Reader`/`io.ReadCloser`.

### Optional Capabilities (Extensions)
Providers can optionally implement these interfaces. The `storage-bridge` engine will type-assert the `Provider` to see if it supports them.

- **`Copier` (Server-Side Copy)**: 
  - Interface: `Copy(ctx, src, dst string) error`
  - Fallback: Engine downloads (`Get`) and uploads (`Put`).
- **`Mover` (Server-Side Move)**:
  - Interface: `Move(ctx, src, dst string) error`
  - Fallback: Engine does `Copy()` then `Delete()`.
- **`MultipartUploader`**:
  - Interface: `CreateMultipart`, `UploadPart`, `CompleteMultipart`
  - Fallback: Engine uses standard `Put`.
- **`Sharer` (Pre-signed URLs)**:
  - Interface: `Share(ctx, path string, expiry time.Duration) (string, error)`
  - Fallback: None. Fails if unsupported.

### Excluded Capabilities
- **`Mkdir` / Directories**: Excluded. Storage is treated as a flat key-value namespace.
- **`Quota`**: Excluded. Too provider-specific.
- **`Resumable`**: Excluded from core. Complex to implement consistently across providers.

By using Go's optional interface type-assertion (like rclone's `fs.MimeTyper`), `storage-bridge` can remain incredibly lightweight while still taking advantage of provider-specific optimizations like zero-cost server-side copies.
