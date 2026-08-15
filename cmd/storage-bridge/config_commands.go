package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	
	"github.com/storage-bridge/core/pkg/config"
	"github.com/storage-bridge/core/pkg/providers/drive"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with a storage service",
}

var authLoginCmd = &cobra.Command{
	Use:   "login [service]",
	Short: "Login to a storage service (e.g. google)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		
		cfgMgr, err := config.NewManager()
		if err != nil {
			return err
		}

		switch service {
		case "google":
			tok, err := drive.AuthLogin(context.Background())
			if err != nil {
				return err
			}
			
			b, err := json.Marshal(tok)
			if err != nil {
				return err
			}
			cfgMgr.Data.Auths["google"] = b
			if err := cfgMgr.Save(); err != nil {
				return err
			}
			fmt.Println("Successfully authenticated and saved token for google!")
		default:
			return fmt.Errorf("unsupported auth service: %s", service)
		}
		
		return nil
	},
}

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage storage providers",
}

var providerAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new named provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		pType, _ := cmd.Flags().GetString("type")
		account, _ := cmd.Flags().GetString("account")
		upstreams, _ := cmd.Flags().GetString("upstreams")
		policy, _ := cmd.Flags().GetString("policy")
		bucket, _ := cmd.Flags().GetString("bucket")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		
		if pType == "" {
			return fmt.Errorf("--type is required")
		}
		
		cfgMgr, err := config.NewManager()
		if err != nil {
			return err
		}
		
		cfgMgr.Data.Providers[name] = config.ProviderConfig{
			Type: pType,
			Params: map[string]string{
				"account":   account,
				"upstreams": upstreams,
				"policy":    policy,
				"bucket":    bucket,
				"endpoint":  endpoint,
			},
		}
		
		if err := cfgMgr.Save(); err != nil {
			return err
		}
		fmt.Printf("Successfully added provider '%s'!\n", name)
		return nil
	},
}

var providerRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a named provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfgMgr, err := config.NewManager()
		if err != nil {
			return err
		}
		
		if _, exists := cfgMgr.Data.Providers[name]; !exists {
			return fmt.Errorf("provider '%s' not found", name)
		}
		
		delete(cfgMgr.Data.Providers, name)
		
		if cfgMgr.Data.DefaultProvider == name {
			cfgMgr.Data.DefaultProvider = ""
		}
		
		if err := cfgMgr.Save(); err != nil {
			return err
		}
		fmt.Printf("Successfully removed provider '%s'\n", name)
		return nil
	},
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all named providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgMgr, err := config.NewManager()
		if err != nil {
			return err
		}
		fmt.Println("Configured providers:")
		for name, pConf := range cfgMgr.Data.Providers {
			fmt.Printf("- %s (Type: %s)\n", name, pConf.Type)
		}
		return nil
	},
}

var providerShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show details of a named provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfgMgr, err := config.NewManager()
		if err != nil {
			return err
		}
		
		pConf, exists := cfgMgr.Data.Providers[name]
		if !exists {
			return fmt.Errorf("provider '%s' not found", name)
		}
		
		b, err := json.MarshalIndent(pConf, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("Provider: %s\n%s\n", name, string(b))
		return nil
	},
}

var providerTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Test a named provider's connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		provider, _, err := resolveProvider(name + ":/")
		if err != nil {
			return fmt.Errorf("✗ Failed to resolve provider: %v", err)
		}
		
		_, err = provider.List(context.Background(), "/")
		if err != nil {
			return fmt.Errorf("✗ Connection test failed: %v", err)
		}
		
		fmt.Printf("✓ Provider '%s' is working correctly!\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	
	providerAddCmd.Flags().StringP("type", "t", "", "Type of provider (s3, drive, local, memory, union)")
	providerAddCmd.Flags().StringP("account", "a", "", "Account identifier (for drive)")
	providerAddCmd.Flags().StringP("upstreams", "u", "", "Comma separated list of upstream providers (for union)")
	providerAddCmd.Flags().StringP("policy", "p", "first", "Policy for routing operations (for union)")
	providerAddCmd.Flags().StringP("bucket", "b", "", "Bucket name (for s3)")
	providerAddCmd.Flags().StringP("endpoint", "e", "", "Custom endpoint URL (for s3, enables R2/B2/Spaces support)")
	
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerAddCmd)
	providerCmd.AddCommand(providerRemoveCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerShowCmd)
	providerCmd.AddCommand(providerTestCmd)
}
