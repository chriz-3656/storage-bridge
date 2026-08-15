package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	storagebridgeconfig "github.com/storage-bridge/core/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "storage-bridge",
	Short: "A lightweight universal storage engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgMgr, err := storagebridgeconfig.NewManager()
		if err != nil || len(cfgMgr.Data.Providers) == 0 {
			fmt.Println("Storage Bridge")
			fmt.Println("One binary. Multiple storage providers.\n")
			fmt.Println("No storage provider is configured.\n")
			fmt.Println("Get started:\n")
			fmt.Println("  storage-bridge login\n")
			fmt.Println("Or:\n")
			fmt.Println("  storage-bridge --help")
			return nil
		}
		return runTUI()
	},
}

func main() {
	initSimpleCommands()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
