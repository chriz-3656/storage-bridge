package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/storage-bridge/core/pkg/storage"
)

var syncWorkers int

var syncCmd = &cobra.Command{
	Use:   "sync [src] [dest]",
	Short: "Synchronize two locations concurrently",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcTarget := resolveSimpleTarget(args[0], "")
		destTarget := resolveSimpleTarget(args[1], "")
		
		srcProv, srcPath, err := resolveProvider(srcTarget)
		if err != nil {
			return fmt.Errorf("failed to resolve source: %v", err)
		}
		
		destProv, destPath, err := resolveProvider(destTarget)
		if err != nil {
			return fmt.Errorf("failed to resolve destination: %v", err)
		}
		
		return runSync(context.Background(), srcProv, srcPath, destProv, destPath)
	},
}

func init() {
	syncCmd.Flags().IntVarP(&syncWorkers, "workers", "w", 4, "Number of concurrent workers")
	rootCmd.AddCommand(syncCmd)
}

func runSync(ctx context.Context, srcProv storage.Provider, srcPath string, destProv storage.Provider, destPath string) error {
	fmt.Printf("Building destination file list for %s:/%s...\n", destProv.Name(), destPath)
	destFiles := make(map[string]*storage.Entry)
	destIter, err := destProv.List(ctx, destPath)
	if err == nil {
		for {
			entry, err := destIter.Next(ctx)
			if err == io.EOF {
				break
			}
			if err != nil {
				// Don't fail entirely, destination might be empty or missing
				break
			}
			if !entry.IsDir {
				destFiles[entry.Path] = entry
			}
		}
	}
	
	fmt.Printf("Building source file list for %s:/%s...\n", srcProv.Name(), srcPath)
	srcIter, err := srcProv.List(ctx, srcPath)
	if err != nil {
		return fmt.Errorf("error listing source: %v", err)
	}
	
	jobs := make(chan *storage.Entry, 100)
	var wg sync.WaitGroup
	errCh := make(chan error, syncWorkers)
	
	// Start workers
	for i := 0; i < syncWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for entry := range jobs {
				// Calculate corresponding destination path.
				relPath := strings.TrimPrefix(entry.Path, srcPath)
				relPath = strings.TrimPrefix(relPath, "/")
				
				targetPath := destPath
				if targetPath != "" && !strings.HasSuffix(targetPath, "/") {
					targetPath += "/"
				}
				targetPath += relPath
				targetPath = strings.TrimPrefix(targetPath, "/")
				
				// Check if it exists in dest and is identical
				destEntry, exists := destFiles[targetPath]
				if exists && destEntry.Size == entry.Size {
					// Skip if identical size
					fmt.Printf("Skipping %s (identical size)\n", relPath)
					continue
				}
				
				fmt.Printf("[Worker %d] Syncing %s → %s\n", workerID, entry.Path, targetPath)
				rc, err := srcProv.Get(ctx, entry.Path, 0, -1)
				if err != nil {
					errCh <- fmt.Errorf("failed to get %s: %v", entry.Path, err)
					return // stop this worker on error
				}
				
				err = destProv.Put(ctx, targetPath, rc, entry.Size, entry.ModTime)
				rc.Close()
				if err != nil {
					errCh <- fmt.Errorf("failed to put %s: %v", targetPath, err)
					return // stop this worker on error
				}
				fmt.Printf("[Worker %d] ✓ Synced %s\n", workerID, relPath)
			}
		}(i)
	}
	
	// Feed jobs
	feedErrCh := make(chan error, 1)
	go func() {
		defer close(jobs)
		for {
			entry, err := srcIter.Next(ctx)
			if err == io.EOF {
				break
			}
			if err != nil {
				feedErrCh <- err
				return
			}
			if !entry.IsDir {
				jobs <- entry
			}
		}
		feedErrCh <- nil
	}()
	
	// Wait for feed to finish
	if err := <-feedErrCh; err != nil {
		return fmt.Errorf("error feeding jobs: %v", err)
	}
	
	// Wait for all workers
	wg.Wait()
	close(errCh)
	
	// Check if any worker encountered an error
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	
	fmt.Println("Sync completed successfully.")
	return nil
}
