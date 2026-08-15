package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	storagebridgeconfig "github.com/storage-bridge/core/pkg/config"
	"github.com/storage-bridge/core/pkg/storage"
)

// runTUI is called when storage-bridge is executed without arguments.
func runTUI() error {
	// Parse targets from SB_TARGETS environment variable
	// Default to "local:." and "memory:/" if empty
	envTargets := os.Getenv("SB_TARGETS")
	var targets []string
	if envTargets != "" {
		targets = strings.Split(envTargets, " ")
	} else {
		cfgMgr, err := storagebridgeconfig.NewManager()
		if err == nil && len(cfgMgr.Data.Providers) > 0 {
			for name := range cfgMgr.Data.Providers {
				targets = append(targets, name+":/")
			}
		} else {
			fmt.Println("No cloud providers connected.\n\nRun:\n\n  storage-bridge login")
			return nil
		}
	}

	providers := make(map[string]storage.Provider)
	providerPaths := make(map[string]string)

	for _, t := range targets {
		if t == "" {
			continue
		}
		p, path, err := resolveProvider(t)
		if err != nil {
			fmt.Printf("Error resolving target %s: %v\n", t, err)
			continue
		}
		// Using the original target string as the unique key
		providers[t] = p
		providerPaths[t] = path
	}

	if len(providers) == 0 {
		fmt.Println("No valid providers to monitor. Set SB_TARGETS environment variable.")
		return nil
	}

	// Main TUI loop
	for {
		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make(map[string]int64)
		errorsMap := make(map[string]error)

		for t, p := range providers {
			wg.Add(1)
			go func(t string, p storage.Provider, path string) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				size, err := p.SpaceUsed(ctx, path)
				mu.Lock()
				if err != nil {
					errorsMap[t] = err
				} else {
					results[t] = size
				}
				mu.Unlock()
			}(t, p, providerPaths[t])
		}

		wg.Wait()

		// Clear screen
		fmt.Print("\033[H\033[2J")

		// Print Header
		fmt.Println("\033[36m" + `
   _____ ____ 
  / ___// __ )
  \__ \/ __  |
 ___/ / /_/ / 
/____/_____/  
Storage Bridge Minimal TUI
` + "\033[0m")
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("%-30s | %-20s\n", "Provider Target", "Storage Used")
		fmt.Println(strings.Repeat("-", 60))

		var total int64
		for _, t := range targets {
			if _, ok := providers[t]; !ok {
				continue
			}
			if err, hasErr := errorsMap[t]; hasErr {
				fmt.Printf("%-30s | \033[31mError: %v\033[0m\n", t, err)
			} else {
				size := results[t]
				total += size
				fmt.Printf("%-30s | %s\n", t, formatBytes(size))
			}
		}

		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("\033[32m%-30s | %s\033[0m\n", "TOTAL COMBINED SPACE", formatBytes(total))
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println("Press Ctrl+C to exit. (Updates every 3 seconds)")

		time.Sleep(3 * time.Second)
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
