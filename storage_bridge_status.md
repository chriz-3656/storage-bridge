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
- ✅ **Amazon S3** (`s3`): Fully integrates with `aws-sdk-go-v2`. Folders are implemented as 0-byte objects, and moves are executed via server-side `CopyObject` + `DeleteObject` to save bandwidth.
- ✅ **Google Drive** (`drive`): Integrates with Google APIs via an automatic local web-server OAuth flow, featuring embedded client secrets. Folder creation and file moving are natively executed via Google Drive parent-ID mutations.

## 3. Configuration & Auth System
- ✅ **Config Manager**: Global persistent JSON configuration is stored at `~/.config/storage-bridge/config.json`.
- ✅ **Stateful Navigation**: Tracks the user's current working directory (`DefaultProviderCwd`) directly inside the config file. This allows users to "cd" into a cloud folder, and have all subsequent uploads and downloads natively default to that specific sub-folder context.
- ✅ **Secure Token Storage**: OAuth tokens (e.g. Google Drive refresh tokens) are securely marshaled into the global config for completely seamless recurring logins.
- ✅ **Provider Aliasing**: Users can assign custom named aliases (e.g. `gdrive`) to specific provider configurations.

## 4. Command Line Interface (CLI)
A robust dual-mode CLI is implemented using the `spf13/cobra` framework, offering both beginner-friendly abstractions and advanced precision.

**Level 1: Everyday Users (Simple Mode)**
Operates purely on the `Default Provider` configured in the system. Path resolution respects the stateful current working directory.
- `login`: Interactive cloud onboarding.
- `logout`: Clears active session.
- `status`: Show connected providers.
- `pwd`: Print the current working directory in the active cloud.
- `cd <directory>`: Traverse into a cloud folder (supports relative paths and root `/`).
- `list [path]`: List files.
- `mkdir <folder>`: Create a new cloud folder.
- `upload <file>`: Upload local file to the current cloud directory.
- `download <file> [dest]`: Download remote file to local disk.
- `move <src> <dest>`: Rename or move a file in the cloud.
- `remove <file>`: Delete file.

**Level 2: Configuration & Power Users**
- `providers`, `default set <provider>`, `provider add <name> --type=<type>`, `provider show <name>`, `provider test <name>`, `provider remove <name>`
- `auth login <account>`

**Level 3: Advanced/Automation (Explicit Target Mode)**
- `ls [provider:target]`, `put [src] [provider:target]`, `get [provider:target] [dest]`, `rm [provider:target]`, `stat [provider:target]`, `cat [provider:target]`

## 5. Development Infrastructure (CI/CD)
- ✅ **Automated GitHub Actions Build**: The binary is compiled continuously in the cloud for multiple architectures (Linux, macOS, Windows).
- ✅ **Cloud-First Workflow**: Development strictly follows an immutable workflow—code is edited, committed, pushed to GitHub, built in the cloud, and the resulting artifact is tested locally, completely bypassing local compilation inconsistencies.

## 6. Challenges Faced & Resolved
- **The Cobra Command Conflict**: During the implementation of the `mkdir` and `move` commands, an architectural conflict arose because both the Advanced UX (Level 3) and the Simple UX (Level 1) attempted to register a command exactly named `mkdir` to the root router. Cobra silently allowed the Advanced command to overwrite the Simple command, causing unexpected schema validation errors for the user (`Error: invalid target format. Expected provider:path`). 
  *Resolution*: We successfully audited the initialization tree and decoupled the Advanced and Simple command roots to strictly preserve Simple UX routing.
- **The Empty Diff Error**: Automated deployment of the `local` provider's `Mkdir` code silently failed to write to disk. 
  *Resolution*: Caught the omission via cloud build logs (which threw a missing interface implementation error), and successfully hotfixed the local provider via absolute patch injection.
- **The Shell REPL Experiment**: To provide a seamless interactive prompt for users (e.g. `demo@mx1[storage-bridge] $`), we developed a full interactive subprocess-isolated Read-Eval-Print Loop (REPL) using Go's `bufio.Scanner` and an internal POSIX-compliant quote parser. We successfully built and deployed it, but upon user testing, we mutually decided the feature was unnecessary feature-creep.
  *Resolution*: Swiftly and cleanly reverted the code out of `main` to keep the CLI perfectly minimal and focused on its core design philosophy.

## 7. Model Context Protocol (MCP) Server
`storage-bridge` exposes a built-in JSON-RPC server over STDIO designed specifically for integration with AI Agents.
**Command**: `storage-bridge mcp`
**Exposed Tools**: `read_file`, `write_file`, `list_directory`

## Summary & Next Steps
The foundation is fully built and highly polished! You have successfully laid down the core interface, integrated robust local and cloud storage (S3 and Google Drive), fully implemented native folder creation, stateful `cd` navigation, cloud-native file moving, and established a fully automated GitHub Actions CI/CD deployment pipeline.

**Next Major Priority**:
- **Phase 6: The Routing / Union Engine**. Now that the providers are fully functional individually, the engine must support combining multiple providers together (e.g., Unioning local and cloud storage into a single directory, or dynamically load-balancing uploads based on available free space).
