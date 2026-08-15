package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	storagebridgeconfig "github.com/storage-bridge/core/pkg/config"
	"github.com/storage-bridge/core/pkg/providers/drive"
)

func resolveSimpleTarget(target string, providerOverride string) string {
	if strings.Contains(target, ":") {
		return target
	}
	cfgMgr, err := storagebridgeconfig.NewManager()

	defaultProv := ""
	if err == nil {
		defaultProv = cfgMgr.Data.DefaultProvider
	}

	if providerOverride != "" {
		defaultProv = providerOverride
	}

	if defaultProv == "" {
		return target
	}

	path := target
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return defaultProv + ":" + path
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Connect a storage provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Storage Bridge")
		fmt.Println("\nNo storage providers are connected yet.")
		fmt.Println("\nConnect a storage provider:")
		fmt.Println("  1. Google Drive")
		fmt.Println("  2. Dropbox")
		fmt.Println("  3. Amazon S3")
		fmt.Println("  4. Cloudflare R2")
		fmt.Print("\nSelect [1-4]: ")

		var selection string
		fmt.Scanln(&selection)

		if selection != "1" {
			return fmt.Errorf("only Google Drive (1) is currently supported in simple login")
		}

		cfgMgr, err := storagebridgeconfig.NewManager()
		if err != nil {
			return err
		}

		tok, err := drive.AuthLogin(context.Background())
		if err != nil {
			return fmt.Errorf("login failed: %v\n\nError code: AUTH_FAILED", err)
		}

		b, err := json.Marshal(tok)
		if err != nil {
			return err
		}

		cfgMgr.Data.Auths["google"] = b
		cfgMgr.Data.Providers["google-drive"] = storagebridgeconfig.ProviderConfig{
			Type:   "drive",
			Params: map[string]string{"account": "google"},
		}
		cfgMgr.Data.DefaultProvider = "google-drive"

		if err := cfgMgr.Save(); err != nil {
			return err
		}

		fmt.Println("\n✓ Google Drive connected\n")
		fmt.Println("Account: authenticated")
		fmt.Println("Provider: google-drive")
		fmt.Println("\nDefault storage provider set to: google-drive\n")
		fmt.Println("You can now use:")
		fmt.Println("  storage-bridge list")
		fmt.Println("  storage-bridge upload <file>")
		fmt.Println("  storage-bridge download <file>")

		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgMgr, err := storagebridgeconfig.NewManager()
		if err != nil {
			return err
		}
		cfgMgr.Data.DefaultProvider = ""
		fmt.Println("Logged out successfully.")
		return cfgMgr.Save()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current status",
	RunE: func(cmd *cobra.Command, args []string) error {
		isJson, _ := cmd.Flags().GetBool("json")
		cfgMgr, err := storagebridgeconfig.NewManager()

		if err != nil || len(cfgMgr.Data.Providers) == 0 {
			if isJson {
				fmt.Println(`{"status": "disconnected"}`)
			} else {
				fmt.Println("Storage Bridge\n\nNo storage provider is configured.")
			}
			return nil
		}

		if isJson {
			out := map[string]string{
				"provider": cfgMgr.Data.DefaultProvider,
				"status":   "connected",
			}
			b, _ := json.Marshal(out)
			fmt.Println(string(b))
			return nil
		}

		fmt.Println("Storage Bridge\n")
		fmt.Printf("Default provider: %s\n", cfgMgr.Data.DefaultProvider)
		fmt.Println("Status: connected\n")
		fmt.Println("Available providers:")
		for name := range cfgMgr.Data.Providers {
			check := " "
			if name == cfgMgr.Data.DefaultProvider {
				check = "✓"
			}
			fmt.Printf("  %s %s\n", check, name)
		}
		fmt.Println("    local")
		fmt.Println("    memory")
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List files (simple)",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		target = resolveSimpleTarget(target, "")
		
		isJson, _ := cmd.Flags().GetBool("json")
		if isJson {
			lsCmd.Flags().Set("json", "true")
		}
		
		return lsCmd.RunE(cmd, []string{target})
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload <file> [destination]",
	Short: "Upload a file (simple)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		source := args[0]

		var target string
		if len(args) > 1 {
			target = args[1]
		} else {
			target = filepath.Base(source)
		}

		target = resolveSimpleTarget(target, provider)

		fmt.Printf("Uploading %s → %s\n\n", filepath.Base(source), target)

		err := putCmd.RunE(cmd, []string{source, target})
		if err != nil {
			return fmt.Errorf("✗ Upload failed\n\nError: %v", err)
		}

		fmt.Println("\n✓ Upload complete")
		fmt.Println(target)
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download <file> [destination]",
	Short: "Download a file (simple)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := resolveSimpleTarget(args[0], "")

		var target string
		if len(args) > 1 {
			target = args[1]
		} else {
			target = "./" + filepath.Base(args[0])
		}

		fmt.Printf("Downloading %s → %s\n\n", source, target)

		err := getCmd.RunE(cmd, []string{source, target})
		if err != nil {
			return fmt.Errorf("✗ Download failed\n\nError: %v", err)
		}

		fmt.Println("\n✓ Download complete")
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove <file>",
	Short: "Remove a file (simple)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := resolveSimpleTarget(args[0], "")
		yes, _ := cmd.Flags().GetBool("yes")

		if !yes {
			fmt.Printf("Delete:\n\n%s\n\nContinue? [y/N]: ", target)
			var resp string
			fmt.Scanln(&resp)
			if strings.ToLower(resp) != "y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		err := rmCmd.RunE(cmd, []string{target})
		if err != nil {
			return fmt.Errorf("✗ Remove failed: %v", err)
		}
		fmt.Println("✓ Removed successfully")
		return nil
	},
}

var catCmd = &cobra.Command{
	Use:   "cat <file>",
	Short: "Print file contents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := resolveSimpleTarget(args[0], "")
		provider, path, err := resolveProvider(target)
		if err != nil {
			return err
		}
		
		// Attempt to detect if it's binary by looking at extension to avoid terminal spam
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".exe" || ext == ".dll" || ext == ".jpg" || ext == ".png" || ext == ".pdf" || ext == ".zip" || ext == ".tar" || ext == ".gz" {
			return fmt.Errorf("file appears to be binary. Use 'download' instead of 'cat'")
		}
		
		rc, err := provider.Get(context.Background(), path, 0, -1)
		if err != nil {
			return err
		}
		defer rc.Close()
		
		_, err = io.Copy(os.Stdout, rc)
		return err
	},
}

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List all providers (Level 2)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgMgr, err := storagebridgeconfig.NewManager()
		if err != nil {
			return err
		}
		fmt.Println("Available providers:")
		for name := range cfgMgr.Data.Providers {
			fmt.Println("- " + name)
		}
		return nil
	},
}

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Manage default provider",
}

var defaultSetCmd = &cobra.Command{
	Use:   "set <provider>",
	Short: "Set the default provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgMgr, err := storagebridgeconfig.NewManager()
		if err != nil {
			return err
		}
		cfgMgr.Data.DefaultProvider = args[0]
		if err := cfgMgr.Save(); err != nil {
			return err
		}
		fmt.Printf("Default provider set to: %s\n", args[0])
		return nil
	},
}

func initSimpleCommands() {
	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	
	uploadCmd.Flags().String("provider", "", "Override default provider")
	removeCmd.Flags().Bool("yes", false, "Skip confirmation")
	
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(providersCmd)
	rootCmd.AddCommand(catCmd)
	
	defaultCmd.AddCommand(defaultSetCmd)
	rootCmd.AddCommand(defaultCmd)
}
