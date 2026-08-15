# Provider Management Research

This document outlines how `rclone` handles provider management, focusing on concepts relevant for commands like `provider list`, `provider add`, `provider remove`.

## Key Concepts in rclone

### Remote Creation
Remotes are created via the `CreateRemote` function. It deletes any old configuration for the given name and sets the `type` before updating remaining values.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L637) (`CreateRemote` function)

### Remote Listing
`rclone` retrieves a list of remotes using `GetRemotes()`, which checks both environment variables (`RCLONE_CONFIG_{NAME}_TYPE`) and the loaded configuration file sections.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L464) (`GetRemotes` function)
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L503) (`GetRemoteNames` function)

### Remote Configuration
Configuration updates are handled by `UpdateRemote`. This applies key-value changes to a remote, handling password obfuscation via the `obscure` package.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L630) (`UpdateRemote` function)

### Remote Deletion
A remote is deleted by removing its section from the config storage using `DeleteSection`.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L86) (`DeleteSection` interface method)
- Source: [fs/config/configfile/configfile.go](file:///home/demo/rclone/fs/config/configfile/configfile.go#L233) (`DeleteSection` implementation)

### Provider Type and Credentials
The provider type is stored as `type` in the remote configuration. Passwords are obfuscated (encrypted) and stored. For OAuth, tokens and secrets are stored using constants like `ConfigToken`, `ConfigClientID`, `ConfigClientSecret`.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L35-L42) (Config Constants)
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L450) (`Remote` struct)

## Extracted Concepts for storage-bridge

For `storage-bridge`, we need:
- `provider list`: A simple aggregation from config storage (similar to `GetRemotes`).
- `provider add`: Initialize a section with `type` and credentials.
- `provider remove`: Drop the section from the config store entirely.
- Auth / Lifecycle: Support OAuth flow and simple token saving without copying the entire interactive UI system of rclone.
