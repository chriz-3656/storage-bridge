# Storage Bridge: Current Project Status

This document outlines the current state of the `storage-bridge` project as of the latest analysis. It details the implemented features, interfaces, and architecture components.

## 1. Core Architecture
The core engine has established a unified storage abstraction (`storage.Provider`) that allows different storage backends to be queried and modified using a single interface.

**Supported Operations (`pkg/storage/provider.go`)**:
- `Stat`: View metadata for a file or directory.
- `List`: Iterate over the contents of a directory.
- `Get`: Download a file stream with optional offset and length.
- `Put`: Upload a file stream with size and modification time.
- `Remove`: Delete a file or directory.
- `SpaceUsed`: Query the total storage utilized by a specific target path.

## 2. Storage Providers
The backend architecture supports pluggable providers. Connection targets are parsed using a unified `provider:path` string format (e.g., `local:./docs`, `s3:my-bucket/path`), or via default routing via the simplified UX.

**Currently Implemented**:
- ✅ **Local Filesystem** (`local`): Reads and writes directly to the OS filesystem.
- ✅ **In-Memory** (`memory`): Ephemeral storage entirely in RAM.
- ✅ **Amazon S3** (`s3`): Fully integrates with `aws-sdk-go-v2` to support cloud object storage.
- ✅ **Google Drive** (`drive`): Integrates with Google APIs via an automatic local web-server OAuth flow, featuring embedded client secrets.

**Pending/Roadmap**:
- ⏳ Dropbox
- ⏳ Backblaze B2

## 3. Configuration & Auth System
- ✅ **Config Manager**: Global persistent JSON configuration is stored at `~/.config/storage-bridge/config.json`.
- ✅ **Secure Token Storage**: OAuth tokens (e.g. Google Drive refresh tokens) are securely marshaled into the global config for completely seamless recurring logins.
- ✅ **Provider Aliasing**: Users can assign custom named aliases (e.g. `gdrive`) to specific provider configurations.

## 4. Command Line Interface (CLI)
A robust dual-mode CLI is implemented using the `spf13/cobra` framework, offering both beginner-friendly abstractions and advanced precision.

**Level 1: Everyday Users (Simple Mode)**
Operates purely on the `Default Provider` configured in the system.
- `login`: Interactive cloud onboarding.
- `logout`: Clears active session.
- `list [path]`: List files.
- `upload <file> [dest]`: Upload local file.
- `download <file> [dest]`: Download remote file.
- `remove <file>`: Delete file.
- `status`: Show connected providers.

**Level 2: Configuration & Power Users**
- `providers`, `default set <provider>`, `provider show <name>`, `provider test <name>`, `provider remove <name>`

**Level 3: Advanced/Automation (Explicit Target Mode)**
- `ls [provider:target]`, `put [src] [provider:target]`, `get [provider:target] [dest]`, `rm [provider:target]`, `stat [provider:target]`, `cat [provider:target]`

## 5. Minimal Text User Interface (TUI)
Running the binary without any subcommands (and with at least one provider configured) automatically launches a minimal TUI dashboard. 
- Monitors the total space used across all connected providers.
- Refreshes automatically every 3 seconds.

## 6. Model Context Protocol (MCP) Server
`storage-bridge` exposes a built-in JSON-RPC server over STDIO designed specifically for integration with AI Agents.

**Command**: `storage-bridge mcp`
**Exposed Tools**:
- `read_file`: Fetches content via `Get` and returns text.
- `write_file`: Streams text content via `Put` into the target provider.
- `list_directory`: Invokes `List` and formats the output cleanly for AI consumption.

## Summary
The foundation is fully built and highly polished! You have successfully laid down the core interface, integrated robust local and cloud storage (S3 and Google Drive), implemented a completely transparent configuration/authentication engine, created a beautiful dual-level CLI, and added modern AI/Agent compatibility via MCP. 

**Next Major Priorities (Based on Roadmap)**:
1. Advanced routing engine / Union providers (Combining targets, fallback behaviors, transparent load balancing).
2. Additional cloud providers (e.g., Dropbox).
