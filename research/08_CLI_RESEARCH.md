# CLI Research

## Architecture and Command Registration
- Rclone uses the `cobra` (`github.com/spf13/cobra`) and `pflag` (`github.com/spf13/pflag`) libraries for command routing and argument parsing. (Source: `cmd/cmd.go`)
- The main entrypoint is `rclone.go`, which simply calls `cmd.Main()`. (Source: `rclone.go`, Function: `main`)
- Commands are split into their own packages under the `cmd/` directory (e.g., `cmd/copy`, `cmd/ls`).
- All active commands are imported blindly via side-effect imports (`_ "github.com/rclone/rclone/cmd/..."`) inside `cmd/all/all.go`. (Source: `cmd/all/all.go`)
- Command registration happens in each command's `init()` function by calling `cobra.Command.AddCommand` (Standard cobra pattern).

## Argument Parsing
- Flags are defined globally in `cmd/cmd.go` or within specific command files using `pflag` (e.g., `flags.StringP`).
- Rclone uses helper functions like `NewFsSrc` and `NewFsDir` to parse positional arguments into filesystem interfaces. (Source: `cmd/cmd.go`, Functions: `NewFsSrc`, `NewFsDir`, `NewFsFile`)

## Errors and Exit Codes
- Rclone has central error handling functions, wrapping standard errors. For example, `fs.Fatalf`, `fs.CountError`. (Source: `cmd/cmd.go`)
- Custom exit codes are defined in `lib/exitcode` (Source: `cmd/cmd.go` imports `github.com/rclone/rclone/lib/exitcode`).
- Standard pre-defined command errors are used, like `errorCommandNotFound = errors.New("command not found")`. (Source: `cmd/cmd.go`)

## Storage-Bridge Recommendation
For `storage-bridge`, we can replicate the `cobra` + `pflag` structure, but with a much smaller set of subcommands (`get`, `put`, `ls`, `search`, `cp`, `provider add`, `auth login`, `route list`). We do not need the complex global flag registry of rclone.
