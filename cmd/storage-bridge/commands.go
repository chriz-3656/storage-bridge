package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	
	"github.com/aws/aws-sdk-go-v2/config"
	s3sdk "github.com/aws/aws-sdk-go-v2/service/s3"
	
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	driveapi "google.golang.org/api/drive/v3"

	storagebridgeconfig "github.com/storage-bridge/core/pkg/config"
	"github.com/storage-bridge/core/pkg/providers/drive"
	"github.com/storage-bridge/core/pkg/providers/local"
	"github.com/storage-bridge/core/pkg/providers/memory"
	"github.com/storage-bridge/core/pkg/providers/s3"
	"github.com/storage-bridge/core/pkg/storage"
)

var (
	memProvider = memory.New()
)

// resolveProvider parses a connection string like "memory:/path/to/file" or "local:/path/to/file"
// It also checks the global config for named providers like "gdrive:/"
func resolveProvider(target string) (storage.Provider, string, error) {
	parts := strings.SplitN(target, ":", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid target format. Expected provider:path, got %s", target)
	}
	
	providerName := parts[0]
	path := parts[1]
	
	cfgMgr, err := storagebridgeconfig.NewManager()
	if err == nil {
		if pConf, ok := cfgMgr.Data.Providers[providerName]; ok {
			switch pConf.Type {
			case "drive":
				account := pConf.Params["account"]
				tokRaw, ok := cfgMgr.Data.Auths[account]
				if !ok {
					return nil, "", fmt.Errorf("auth account '%s' not found", account)
				}
				var tok oauth2.Token
				if err := json.Unmarshal(tokRaw, &tok); err != nil {
					return nil, "", err
				}
				
				conf := &oauth2.Config{
					ClientID:     drive.DefaultClientID,
					ClientSecret: drive.DefaultClientSecret,
					Endpoint:     google.Endpoint,
					Scopes:       []string{driveapi.DriveScope},
				}
				client := conf.Client(context.Background(), &tok)
				provider, err := drive.New(context.Background(), client)
				if err != nil {
					return nil, "", err
				}
				return provider, path, nil
			}
		}
	}
	
	switch providerName {
	case "memory":
		return memProvider, path, nil
	case "local":
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", err
		}
		provider, err := local.New(cwd)
		if err != nil {
			return nil, "", err
		}
		return provider, path, nil
	case "s3":
		parts := strings.SplitN(path, "/", 2)
		bucket := parts[0]
		key := ""
		if len(parts) > 1 {
			key = parts[1]
		}
		
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, "", err
		}
		client := s3sdk.NewFromConfig(cfg)
		return s3.New(client, bucket), key, nil
	case "drive":
		conf := &oauth2.Config{
			ClientID:     drive.DefaultClientID,
			ClientSecret: drive.DefaultClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{driveapi.DriveScope},
		}
		tokRaw, err := os.ReadFile("token.json")
		if err != nil {
			return nil, "", fmt.Errorf("token.json not found. Run 'auth login google' first")
		}
		var tok oauth2.Token
		json.Unmarshal(tokRaw, &tok)
		
		client := conf.Client(context.Background(), &tok)
		provider, err := drive.New(context.Background(), client)
		if err != nil {
			return nil, "", err
		}
		return provider, path, nil
	default:
		return nil, "", fmt.Errorf("unknown provider: %s", providerName)
	}
}

func init() {
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(statCmd)
	rootCmd.AddCommand(mcpCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls [target]",
	Short: "List a directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, path, err := resolveProvider(args[0])
		if err != nil {
			return err
		}
		
		iter, err := provider.List(context.Background(), path)
		if err != nil {
			return err
		}
		
		for {
			entry, err := iter.Next(context.Background())
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			
			if entry.IsDir {
				fmt.Printf("DIR  %s\n", entry.Path)
			} else {
				fmt.Printf("%10d %s\n", entry.Size, entry.Path)
			}
		}
		return nil
	},
}

var putCmd = &cobra.Command{
	Use:   "put [source] [target]",
	Short: "Upload a file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		target := args[1]
		
		provider, path, err := resolveProvider(target)
		if err != nil {
			return err
		}
		
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		defer f.Close()
		
		info, err := f.Stat()
		if err != nil {
			return err
		}
		
		return provider.Put(context.Background(), path, f, info.Size(), info.ModTime())
	},
}

var getCmd = &cobra.Command{
	Use:   "get [target] [destination]",
	Short: "Download a file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		dest := args[1]
		
		provider, path, err := resolveProvider(target)
		if err != nil {
			return err
		}
		
		rc, err := provider.Get(context.Background(), path, 0, -1)
		if err != nil {
			return err
		}
		defer rc.Close()
		
		f, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer f.Close()
		
		_, err = io.Copy(f, rc)
		return err
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm [target]",
	Short: "Remove a file or directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, path, err := resolveProvider(args[0])
		if err != nil {
			return err
		}
		return provider.Remove(context.Background(), path)
	},
}

var statCmd = &cobra.Command{
	Use:   "stat [target]",
	Short: "Show metadata for a file or directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, path, err := resolveProvider(args[0])
		if err != nil {
			return err
		}
		entry, err := provider.Stat(context.Background(), path)
		if err != nil {
			return err
		}
		fmt.Printf("Path: %s\nIsDir: %v\nSize: %d\nModTime: %s\n", entry.Path, entry.IsDir, entry.Size, entry.ModTime.Format(time.RFC3339))
		return nil
	},
}
