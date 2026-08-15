# Provider Architecture

This document traces how a cloud storage provider is represented, initialized, and used in rclone.

## Provider Interface and Object Abstraction
A provider is fundamentally an implementation of:
1. **`fs.Fs`** (Source: `fs/types.go:17`): The filesystem/bucket representation.
2. **`fs.Object`** (Source: `fs/types.go:82`): The file representation.

## Registration
Providers are self-registering. In the provider's package (e.g., `backend/s3/s3.go`), an `init()` function is defined that calls:
`fs.Register(&fs.RegInfo{...})` (Source: `fs/registry.go:34`).

`fs.RegInfo` defines:
- `Name`: e.g., "s3"
- `Options`: A list of `fs.Option` structs defining the config fields (e.g., access key, secret key, region).
- `NewFs`: The constructor function: `func(ctx context.Context, name string, root string, config configmap.Mapper) (Fs, error)`.

## Initialization Lifecycle
1. **USER**: Invokes a command like `rclone copy /local s3:bucket/path`.
2. **CLI Framework**: Parses the remote `s3:bucket/path` (Source: `cmd.NewFsFile` in `cmd/cmd.go:85`, calling `fspath.SplitFs`).
3. **Provider Configuration**: The config system looks up the `s3` remote in the user's `rclone.conf` and wraps it in a `configmap.Mapper`.
4. **Backend Discovery & Initialization**: Rclone looks up the `s3` provider in the `fs.Registry`. It invokes the `NewFs` function provided in the `RegInfo`.
5. **Authentication**: Inside `NewFs`, the provider reads the auth tokens/keys from the `configmap.Mapper`, constructs an authenticated HTTP client (e.g., using `fs/fshttp`), and verifies access.
6. **Remote Creation**: `NewFs` returns an instance of the provider's `fs.Fs` implementation, pinned to `bucket/path`.
7. **Filesystem Operations**: The `copy` command invokes `fs.Fs.Put` or `fs.Fs.List` to interact with the backend.
8. **Provider API**: The provider implementation translates these interface calls into REST API requests to the cloud storage (e.g., AWS S3 API).

## Capabilities
Not all backends support all features (e.g., some don't support moving files server-side). Backends declare capabilities in two ways:
1. **`fs.Info.Features()`**: Returns an `*fs.Features` struct indicating boolean support for features (e.g., `Copy`, `Move`, `Purge`).
2. **Optional Interfaces**: Go interface type-checking. For example, if an `fs.Object` implements `fs.MimeTyper` (Source: `fs/types.go:159`), rclone knows it can query the MIME type.
