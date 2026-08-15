# Rclone Architecture

Rclone's architecture is highly modular and designed around a core set of abstractions for file systems and objects, allowing it to interface with dozens of cloud storage providers using a unified API.

## Core Packages
- **`fs`** (`fs/`): The absolute core of rclone. It defines the central interfaces (`Fs`, `Object`, `Directory`) and core types used throughout the program.
- **`cmd`** (`cmd/`): The CLI framework based on Cobra. Each subcommand (e.g., `copy`, `sync`) is implemented here.
- **`backend`** (`backend/`): Contains the implementations of the cloud storage providers (e.g., `backend/s3`, `backend/local`).
- **`vfs`** (`vfs/`): The Virtual File System layer, which builds on top of `fs.Fs` to provide a standard file system interface (used by `rclone mount`).

## Filesystem Abstractions
The most important abstraction in rclone is the `Fs` interface.
- **`fs.Fs`** (Source: `fs/types.go:17`): Represents a cloud storage system. It requires implementing `List`, `NewObject`, `Put`, `Mkdir`, and `Rmdir`, plus inheriting `fs.Info`.
- **`fs.Object`** (Source: `fs/types.go:82`): Represents a file. It inherits `fs.ObjectInfo` and adds methods like `SetModTime`, `Open`, `Update`, and `Remove`.
- **`fs.Directory`** (Source: `fs/types.go:136`): Represents a directory.

## Backend/Provider System & Registration
Providers implement the `fs.Fs` interface. They register themselves with the core during initialization.
- **Registration**: Providers call `fs.Register(&fs.RegInfo{...})` in their `init()` functions. 
- **Type**: `fs.RegInfo` (Source: `fs/registry.go:34`). This struct contains the provider's `Name`, `Description`, `NewFs` constructor, and `Options` (configuration schema).
- **Example**: `backend/s3/s3.go` calls `fs.Register` to register the `s3` provider.

## Config System
Configuration is handled dynamically based on the options defined in `fs.RegInfo`.
- **Type**: `configmap.Mapper` is passed to the `NewFs` function.
- **Responsibility**: It abstracts reading configuration from the user's config file (or environment variables/flags) based on the schema defined by the provider's `Options`.

## Auth System
Authentication is usually handled on a per-backend basis since different providers use different mechanisms (OAuth2, API keys, basic auth). Rclone provides utilities in `lib/oauthutil` and HTTP client wrappers in `fs/fshttp` to facilitate this.

## Command/CLI System
- **Framework**: `github.com/spf13/cobra` (Source: `cmd/cmd.go`).
- **Registration**: Commands define a `cobra.Command` and register it (e.g., `cmd.Root.AddCommand`). The `cmd/all/all.go` file acts as an import sink to ensure all commands are compiled into the binary.

## Object Model
- **`fs.DirEntry`** (Source: `fs/types.go:118`): The common subset of `fs.Object` and `fs.Directory`. Returned from directory listings.
- **`fs.ObjectInfo`** (Source: `fs/types.go:104`): Read-only info about an object, adding `Hash` and `Storable` to `DirEntry`.

## Streaming Architecture
- **Upload**: The `fs.Fs.Put` and `fs.Object.Update` methods take an `io.Reader`. If the size is unknown, `src.Size()` returns `-1` (Source: `fs/types.go:42`).
- **Download**: `fs.Object.Open` returns an `io.ReadCloser`. Rclone supports seekable reads through `fs.OpenOption` (like `RangeOption`).

## Error Handling
Rclone defines standardized sentinel errors to ensure consistent behavior across backends.
- **Examples**: `fs.ErrorObjectNotFound`, `fs.ErrorIsDir`, `fs.ErrorIsFile` (Seen in `cmd/cmd.go:94`). Backends translate their native HTTP/API errors into these standard `fs` errors.
