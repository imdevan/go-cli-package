package pipeline

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TagList prints all git tags sorted newest-first, up to limit entries.
func TagList(rootDir string, limit int) error {
	args := []string{"tag", "-l", "--sort=-v:refname"}
	if limit > 0 {
		args = append(args, fmt.Sprintf("--count=%d", limit))
	}

	// Use git tag -l and pipe through head manually for portability
	cmd := exec.Command("git", "tag", "-l", "--sort=-v:refname")
	cmd.Dir = rootDir

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git tag list: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		fmt.Println("No tags found.")
		return nil
	}

	fmt.Printf("📋 Tags (%d):\n", len(lines))
	for i, line := range lines {
		if limit > 0 && i >= limit {
			fmt.Printf("  ... and %d more\n", len(lines)-limit)
			break
		}
		fmt.Printf("  %s\n", line)
	}
	return nil
}

// TagCreate creates an annotated git tag and pushes it to origin.
// If the tag already exists, it returns an error unless force is true,
// in which case the existing tag is deleted and recreated.
func TagCreate(rootDir, version string, force bool) error {
	version = strings.TrimPrefix(version, "v")
	tag := "v" + version

	fmt.Printf("🏷️  Creating tag %s...\n", tag)

	// Check whether the tag already exists
	checkCmd := exec.Command("git", "rev-parse", tag)
	checkCmd.Dir = rootDir
	checkCmd.Stdout = nil
	checkCmd.Stderr = nil

	exists := checkCmd.Run() == nil
	if exists {
		if !force {
			return fmt.Errorf("tag %s already exists; use --force to delete and recreate", tag)
		}
		fmt.Printf("⚠️  Tag %s already exists — deleting and recreating (--force)\n", tag)
		if err := TagDelete(rootDir, version); err != nil {
			return err
		}
	}

	if err := runCmdInDir(rootDir, "git", "tag", "-a", tag, "-m", "Release "+tag); err != nil {
		return fmt.Errorf("git tag create: %w", err)
	}

	if err := runCmdInDir(rootDir, "git", "push", "origin", tag); err != nil {
		return fmt.Errorf("git push tag: %w", err)
	}

	fmt.Printf("✅ Tag %s created and pushed\n", tag)
	return nil
}

// TagDelete removes a git tag locally and from the origin remote.
// Missing local or remote tags are treated as non-fatal.
func TagDelete(rootDir, version string) error {
	version = strings.TrimPrefix(version, "v")
	tag := "v" + version

	fmt.Printf("🗑️  Deleting tag %s...\n", tag)

	// Delete locally (ignore error if not present)
	localDel := exec.Command("git", "tag", "-d", tag)
	localDel.Dir = rootDir
	var localOut bytes.Buffer
	localDel.Stdout = &localOut
	localDel.Stderr = &localOut
	if err := localDel.Run(); err != nil {
		fmt.Printf("   Local tag not found (skipping local delete)\n")
	}

	// Delete from remote (ignore error if not present)
	remoteDel := exec.Command("git", "push", "origin", ":refs/tags/"+tag)
	remoteDel.Dir = rootDir
	var remoteOut bytes.Buffer
	remoteDel.Stdout = &remoteOut
	remoteDel.Stderr = &remoteOut
	if err := remoteDel.Run(); err != nil {
		fmt.Printf("   Remote tag not found (skipping remote delete)\n")
	}

	fmt.Printf("✅ Tag %s deleted\n", tag)
	return nil
}

// resolveVersion returns the provided version string, falling back to the
// most recent git tag, and finally to the config version.
func resolveVersion(rootDir, provided string, cfg *Config) string {
	if provided != "" {
		return provided
	}
	if v := getGitTagVersion(rootDir); v != "" {
		return v
	}
	return cfg.Version
}

// tagExists reports whether the given tag exists in the repository.
func tagExists(rootDir, tag string) bool {
	cmd := exec.Command("git", "rev-parse", tag)
	cmd.Dir = rootDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// currentGitDir returns the root directory of the git repo containing rootDir.
// Falls back to rootDir itself on error.
func currentGitDir(rootDir string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = rootDir
	out, err := cmd.Output()
	if err != nil {
		return rootDir
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return rootDir
	}
	return result
}

// gitRootDir returns the root directory for git operations.
// It prefers the actual git repo root so that tags are created in the correct repo.
func gitRootDir(rootDir string) string {
	if _, err := os.Stat(rootDir + "/.git"); err == nil {
		return rootDir
	}
	return currentGitDir(rootDir)
}
