# Configuration and Auth Research

This document outlines how `rclone` manages configuration, authentication, and credentials securely.

## Configuration Handling in rclone

### Config Storage
rclone stores its configuration in an INI-based file (`rclone.conf`). The `configfile` package manages loading and saving this file.
- Source: [fs/config/configfile/configfile.go](file:///home/demo/rclone/fs/config/configfile/configfile.go#L23) (`Storage` struct)

### Encrypted Configuration
The configuration file can be optionally encrypted. `configfile.go` checks for encryption and passes the config through `Decrypt()` when loading and `Encrypt()` when saving.
- Source: [fs/config/configfile/configfile.go](file:///home/demo/rclone/fs/config/configfile/configfile.go#L80) (`config.Decrypt` call)
- Source: [fs/config/configfile/configfile.go](file:///home/demo/rclone/fs/config/configfile/configfile.go#L148) (`config.Encrypt` call)
- Specific fields (like passwords) are also individually obfuscated using the `obscure` package.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L584) (`obscure.Obscure` call)

### Environment Variables
Environment variables take precedence over config files. `GetValue` first looks up `RCLONE_CONFIG_{REMOTE}_{KEY}` before reading from the loaded data.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L429) (`GetValue` function)

### Platform-Specific Configuration Locations
rclone determines the default configuration file location based on the OS:
- **Windows**: Uses the `%APPDATA%\rclone` directory (e.g. `C:\Users\User\AppData\Roaming\rclone\rclone.conf`)
- **macOS / Linux (Unix)**: Follows the XDG Base Directory specification, defaulting to `$XDG_CONFIG_HOME/rclone` or `~/.config/rclone`.
- **Fallback**: Looks in `~/.rclone.conf` or the current executable's directory.
- Source: [fs/config/config.go](file:///home/demo/rclone/fs/config/config.go#L212-L252) (`makeConfigPath` function)

## Proposed Secure Configuration Model for storage-bridge

For `storage-bridge`, we should implement a model that strictly separates configuration from credentials:
1. **Separation**: Keep general provider settings (types, aliases, paths) in a plain text configuration file (e.g., JSON, YAML, or INI) at the standard OS-specific config location.
2. **Secure Credential Storage**: Do not store passwords, OAuth tokens, or secrets in the plain text config. Instead, use the OS's native secure credential manager (Keyring on Linux, Keychain on macOS, Credential Manager on Windows) or a separate strictly encrypted file for secrets.
3. **Environment Override**: Like rclone, allow environment variables to inject both configuration and credentials for containerized/CI environments.
