package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// @docs-command:
//
//		name: watch
//		description:
//			Monitors source files for changes and automatically re-runs generate.
//			Watched patterns: *.md, *.go, package.toml
//			Excluded paths: node_modules/, docs/src/content/docs/, .git/
//		example:
//			```bash
//			go-cli-docs watch
//			go-cli-docs watch --gen-api-docs=false
//			```
func newWatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch source files and re-generate documentation on change",
		Long: `watch monitors source files for changes and automatically re-runs generate.

Watched patterns: *.md, *.go, package.toml
Excluded paths:   node_modules/, docs/src/content/docs/, .git/

Press Ctrl+C to stop watching.

Run the Astro dev server separately:
  cd docs && bun run dev`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(genAPIDocs)
		},
	}

	return cmd
}

// watchedExts are file extensions that trigger a regeneration.
var watchedExts = map[string]bool{
	".md":   true,
	".go":   true,
	".toml": true,
}

// excludedDirs are directory substrings to ignore.
var excludedDirs = []string{
	"node_modules",
	filepath.Join("docs", "src", "content", "docs"),
	".git",
}

func runWatch(genAPIDocs bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	if err := addWatchDirs(watcher, "."); err != nil {
		return fmt.Errorf("failed to add watch dirs: %w", err)
	}

	// Initial generation.
	fmt.Println("▶  Running initial generate...")
	if err := runGenerate(genAPIDocs); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  generate error: %v\n", err)
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("👀 Watching for changes… (Ctrl+C to stop)")

	for {
		select {
			case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if !isWatchedEvent(event) {
				continue
			}
			ts := time.Now().Format("15:04:05")
			fmt.Printf("[%s] 🔄 %s changed – regenerating...\n", ts, event.Name)
			if err := runGenerate(genAPIDocs); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  generate error: %v\n", err)
			} else {
				fmt.Printf("[%s] ✅ Done\n", ts)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "⚠️  watcher error: %v\n", err)

		case <-quit:
			fmt.Println("\n👋 Stopping watcher.")
			return nil
		}
	}
}

// addWatchDirs recursively adds directories to the watcher, skipping excluded paths.
func addWatchDirs(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		if !d.IsDir() {
			return nil
		}
		if isExcluded(path) {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}

// isWatchedEvent returns true when the event involves a file we care about.
func isWatchedEvent(event fsnotify.Event) bool {
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
		return false
	}
	ext := filepath.Ext(event.Name)
	if !watchedExts[ext] {
		return false
	}
	return !isExcluded(event.Name)
}

// isExcluded returns true when a path falls under an excluded directory.
func isExcluded(path string) bool {
	for _, excl := range excludedDirs {
		if containsPath(path, excl) {
			return true
		}
	}
	return false
}

// containsPath reports whether path contains the sub segment.
func containsPath(path, sub string) bool {
	clean := filepath.Clean(path)
	cleanSub := filepath.Clean(sub)
	// Walk each directory component.
	for p := clean; p != "." && p != "/"; p = filepath.Dir(p) {
		if p == cleanSub || filepath.Base(p) == cleanSub {
			return true
		}
	}
	return false
}
