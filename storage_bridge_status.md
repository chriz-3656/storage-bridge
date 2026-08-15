# Storage Bridge: Current Project Status

This document outlines the current state of the `storage-bridge` project as of the latest analysis. It details the implemented features, architecture components, and the development journey.

## 1. Core Architecture
The core engine has established a unified storage abstraction (`storage.Provider`) that allows different storage backends to be queried and modified using a single interface.

**Supported Operations (`pkg/storage/provider.go`)**:
- `Stat`: View metadata for a file or directory.
- `List`: Iterate over the contents of a directory.
- `Get`: Download a file stream with optional offset and length.
- `Put`: Upload a file stream with size and modification time.
- `Remove`: Delete a file or directory.
- `SpaceUsed`: Query the total storage utilized by a specific target path.
- **`Mkdir`**: Create new folders/directories inside the cloud natively.
- **`Move`**: Rename and move files and directories natively within the cloud.

## 2. Storage Providers
The backend architecture supports pluggable providers. Connection targets are parsed using a unified `provider:path` string format (e.g., `local:./docs`, `s3:my-bucket/path`), or via default routing via the simplified UX.

**Currently Implemented**:
- ✅ **Local Filesystem** (`local`): Reads, writes, creates folders, and moves files directly to the OS filesystem.
- ✅ **In-Memory** (`memory`): Ephemeral storage entirely in RAM with pseudo-folder routing.
- ✅ **Amazon S3** (`s3`): Fully integrates with `aws-sdk-go-v2`. Folders are implemented as 0-byte objects, and moves are executed via server-side `CopyObject` + `DeleteObject` to save bandwidth. Custom endpoints enable R2, B2, and Spaces.
- ✅ **Google Drive** (`drive`): Integrates with Google APIs via an automatic local web-server OAuth flow, featuring embedded client secrets. Folder creation and file moving are natively executed via Google Drive parent-ID mutations.
- ✅ **Union Meta-Provider** (`union`): Combines multiple upstream storage providers and dynamically routes traffic using advanced policies (e.g. `first`, `all`).

## 3. Configuration & Auth System
- ✅ **Config Manager**: Global persistent JSON configuration is stored at `~/.config/storage-bridge/config.json`.
- ✅ **Stateful Navigation**: Tracks the user's current working directory (`DefaultProviderCwd`) directly inside the config file. This allows users to "cd" into a cloud folder, and have all subsequent uploads and downloads natively default to that specific sub-folder context.
- ✅ **Secure Token Storage**: OAuth tokens (e.g. Google Drive refresh tokens) are securely marshaled into the global config for completely seamless recurring logins.
- ✅ **Provider Aliasing**: Users can assign custom named aliases (e.g. `gdrive`) to specific provider configurations.

## 4. Project Roadmap

| Phase | Status | Goal |
|---|---|---|
| Phase 1 | **Complete** | Initialize structure and design interfaces. |
| Phase 2 | **Complete** | Setup CI pipeline and GitHub Actions. |
| Phase 3 | **Complete** | Implement Memory and Local FS providers. |
| Phase 4 | **Complete** | Advanced Cloud Providers (S3 & Google Drive). |
| Phase 5 | **Complete** | Stateful CLI & Architecture (REPL reverted; built persistent Config/Cwd/Simple UX). |
| Phase 6 | **Complete** | Routing Engine / Union Providers (Combining multiple targets). |
| Phase 7 | **Complete** | Advanced Data Sync & Concurrency (Sync engine, diffs, multiprocessing). |
| Phase 8 | **Complete** | Expanding Provider Ecosystem (Universal S3 API endpoint support for Cloudflare R2, Backblaze B2, DigitalOcean Spaces). |

## 5. Command Line Interface (CLI)
A robust dual-mode CLI is implemented using the `spf13/cobra` framework, offering both beginner-friendly abstractions and advanced precision.

**Level 1: Everyday Users (Simple Mode)**
Operates purely on the `Default Provider` configured in the system. Path resolution respects the stateful current working directory.
- `login`, `logout`, `status`, `pwd`, `cd <directory>`, `list [path]`, `mkdir <folder>`, `upload <file>`, `download <file> [dest]`, `move <src> <dest>`, `remove <file>`, `sync [src] [dest]`

**Level 2: Configuration & Power Users**
- `providers`, `default set <provider>`, `provider add <name> --type=<type>`, `provider show <name>`, `provider test <name>`, `provider remove <name>`, `auth login <account>`

**Level 3: Advanced/Automation (Explicit Target Mode)**
- `ls [provider:target]`, `put [src] [provider:target]`, `get [provider:target] [dest]`, `rm [provider:target]`, `stat [provider:target]`, `cat [provider:target]`

## 6. Development Infrastructure (CI/CD)
- ✅ **Automated GitHub Actions Build**: The binary is compiled continuously in the cloud for multiple architectures (Linux, macOS, Windows).
- ✅ **Cloud-First Workflow**: Development strictly follows an immutable workflow—code is edited, committed, pushed to GitHub, built in the cloud, and the resulting artifact is tested locally, completely bypassing local compilation inconsistencies.

## 7. Model Context Protocol (MCP) Server
`storage-bridge` exposes a built-in JSON-RPC server over STDIO designed specifically for integration with AI Agents.
**Command**: `storage-bridge mcp`
**Exposed Tools**: `read_file`, `write_file`, `list_directory`, `make_directory`, `move_file`

## Summary
The foundation is fully built and highly polished! We have successfully laid down the core interface, integrated robust local and cloud storage, fully implemented native folder creation, stateful `cd` navigation, cloud-native file moving, a robust JSON-RPC MCP server, and established a fully automated GitHub Actions CI/CD deployment pipeline. Finally, a parallelized multi-threaded sync engine and a dynamic union load balancer round out the professional capabilities of the suite.
