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
				"account": account,
			},
		}
		
		if err := cfgMgr.Save(); err != nil {
			return err
		}
		fmt.Printf("Successfully added provider '%s'!\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	
	providerAddCmd.Flags().String("type", "", "Provider type (e.g. drive, s3, local)")
	providerAddCmd.Flags().String("account", "", "Auth account name (e.g. google)")
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerAddCmd)
}
