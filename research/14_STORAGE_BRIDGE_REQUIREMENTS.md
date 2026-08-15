# Storage Bridge Requirements

Based on the research into `rclone` and the desired end-state for `storage-bridge`, here are the core requirements and constraints for the project:

## 1. Core Engineering Constraints
- **Language**: Go
- **Distribution**: Single standalone executable
- **Dependencies**: No external runtime requirements (No Python, Node.js, Java, DBs, frontend). 
- **Binary Size**: Optimized and lean. Exclude unused backend SDKs using build tags to prevent bloat.
- **Cross-Platform**: Support Linux (AMD64/ARM64), Windows (AMD64/ARM64), and macOS (AMD64/ARM64).

## 2. Core Operational Requirements
- **Streaming-First**: Zero unnecessary temporary file storage. Zero loading of entire files into RAM. Use efficient `io.Reader` and `io.Writer` streaming paths.
- **Provider Architecture**: Extensible backend interface capable of wrapping Local, S3, Google Drive, Backblaze B2, Dropbox, etc.
- **Capability Detection**: Gracefully handle missing provider capabilities (e.g. if a provider cannot do server-side copy or resumable uploads, fallback or error gracefully).
- **Resilience**: Implement intelligent retries for temporary network errors, rate-limiting backoffs, and interrupted transfers.

## 3. Configuration & Security
- **Strict Separation**: Separate generic configuration (e.g., limits, routing rules) from sensitive credentials (e.g., OAuth tokens, Secret Keys).
- **Standardized Locations**: Use OS-specific standard configuration paths (e.g. `~/.config/storage-bridge` on Linux, `%APPDATA%` on Windows).
- **Environment Overrides**: Support overriding configuration and credentials via environment variables for CI/CD and container usage.

## 4. Interfaces and Access
- **CLI**: A modern, Unix-style CLI with structured commands (`get`, `put`, `ls`, `provider add`, etc.), cross-platform compatibility (PowerShell, CMD, Terminal).
- **MCP/STDIO**: Built-in Model Context Protocol integration over standard input/output for AI agents, allowing them to invoke storage commands seamlessly.
- **Optional HTTP**: Future-proof for a lightweight API server mode, but without a heavy bundled web GUI.

## 5. Management & Routing
- **Provider Management**: Ability to seamlessly list, add, show, and test providers.
- **Lightweight Routing**: Support dynamic routing policies such as priority-based, path-based, round-robin, and fallback/most-available, without the complexity of a full FUSE overlay filesystem.
- **Auto-Update**: Support checking and fetching checksum-verified binary updates.
