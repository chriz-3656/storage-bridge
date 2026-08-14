package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "storage-bridge",
	Short: "A lightweight universal storage engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
