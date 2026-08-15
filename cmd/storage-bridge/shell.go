package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
	storagebridgeconfig "github.com/storage-bridge/core/pkg/config"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start an interactive storage-bridge shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		u, _ := user.Current()
		hostname, _ := os.Hostname()
		
		username := "user"
		if u != nil {
			username = u.Username
		}
		host := "localhost"
		if hostname != "" {
			host = hostname
		}

		fmt.Println("Welcome to the Storage Bridge Shell!")
		fmt.Println("Type 'help' for commands, or 'exit' to quit.")
		
		scanner := bufio.NewScanner(os.Stdin)
		
		for {
			cfgMgr, _ := storagebridgeconfig.NewManager()
			provider := "none"
			cwd := "/"
			if cfgMgr != nil && cfgMgr.Data.DefaultProvider != "" {
				provider = cfgMgr.Data.DefaultProvider
				if cfgMgr.Data.DefaultProviderCwd != "" {
					cwd = cfgMgr.Data.DefaultProviderCwd
				}
			}
			
			// Format: demo@mx1[storage-bridge:google-drive:/test] $
			prompt := fmt.Sprintf("%s@%s[storage-bridge:%s:%s] $ ", username, host, provider, cwd)
			fmt.Print(prompt)
			
			if !scanner.Scan() {
				break
			}
			
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if line == "exit" || line == "quit" {
				break
			}
			
			parts := splitShellArgs(line)
			
			// Execute the command as a subprocess to guarantee perfect state isolation
			execCmd := exec.Command(os.Args[0], parts...)
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			
			_ = execCmd.Run() // Subprocess handles its own error printing
		}
		
		fmt.Println("Goodbye!")
		return nil
	},
}

func splitShellArgs(s string) []string {
	var args []string
	var buf strings.Builder
	inQuote := false
	
	for _, r := range s {
		if r == '"' {
			inQuote = !inQuote
			continue
		}
		if r == ' ' && !inQuote {
			if buf.Len() > 0 {
				args = append(args, buf.String())
				buf.Reset()
			}
			continue
		}
		buf.WriteRune(r)
	}
	if buf.Len() > 0 {
		args = append(args, buf.String())
	}
	return args
}
