package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func checkGitRepo(t *testing.T, dir string, expectedBranch string) {
	t.Helper()
	// Check if git directory exists
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Errorf("expected .git directory to exist in %s: %v", dir, err)
		return
	}

	// Check branch name
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Errorf("failed to check git branch: %v", err)
		return
	}

	actualBranch := strings.TrimSpace(string(out))
	if actualBranch != expectedBranch {
		t.Errorf("expected branch %s, got %s", expectedBranch, actualBranch)
	}

	// Check if there are commits
	cmd = exec.Command("git", "log", "-n", "1", "--oneline")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Errorf("expected at least one commit in %s", dir)
	}
}

func TestInitializeHomebrew(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		Name:        "test-cli",
		PackageName: "test-cli-package",
		Version:     "0.1.0",
		Homepage:    "https://example.com",
		Repository:  "https://github.com/org/test-cli",
		Description: "A generic test CLI",
	}

	pipelines := NewPipelines(tmpDir, cfg)

	err := InitializeHomebrew(pipelines.Homebrew, true)
	if err != nil {
		t.Fatalf("InitializeHomebrew failed: %v", err)
	}

	// Verify formula file existence
	if _, err := os.Stat(pipelines.Homebrew.FormulaPath); err != nil {
		t.Fatalf("expected formula file to exist: %v", err)
	}

	formulaContent, err := os.ReadFile(pipelines.Homebrew.FormulaPath)
	if err != nil {
		t.Fatalf("failed to read formula file: %v", err)
	}

	if !strings.Contains(string(formulaContent), "class TestCliPackage < Formula") {
		t.Errorf("formula class name not capitalized correctly: %s", string(formulaContent))
	}
	if !strings.Contains(string(formulaContent), "desc \"A generic test CLI\"") {
		t.Errorf("formula desc not template substituted correctly")
	}

	// Verify README
	readmePath := filepath.Join(pipelines.Homebrew.TapDir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("expected README to exist: %v", err)
	}

	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}
	if !strings.Contains(string(readmeContent), "brew tap org/test-cli-package") {
		t.Errorf("expected README to contain tap instruction, got: %s", string(readmeContent))
	}

	// Verify Git repository state
	checkGitRepo(t, pipelines.Homebrew.TapDir, "main")
}

func TestInitializeAUR(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		Name:        "test-cli",
		PackageName: "test-cli-package",
		Version:     "0.1.0",
		Homepage:    "https://example.com",
		Repository:  "https://github.com/org/test-cli",
		Description: "A generic test CLI",
		Author:      "test-author@example.com",
	}

	pipelines := NewPipelines(tmpDir, cfg)

	err := InitializeAUR(pipelines.AUR, true)
	if err != nil {
		t.Fatalf("InitializeAUR failed: %v", err)
	}

	// Verify PKGBUILD existence
	if _, err := os.Stat(pipelines.AUR.PKGBUILDPath); err != nil {
		t.Fatalf("expected PKGBUILD to exist: %v", err)
	}

	pkgbuildContent, err := os.ReadFile(pipelines.AUR.PKGBUILDPath)
	if err != nil {
		t.Fatalf("failed to read PKGBUILD: %v", err)
	}
	if !strings.Contains(string(pkgbuildContent), "pkgname=test-cli-package") {
		t.Errorf("expected PKGBUILD to contain pkgname, got: %s", string(pkgbuildContent))
	}
	if !strings.Contains(string(pkgbuildContent), "# Maintainer: test-author@example.com") {
		t.Errorf("expected PKGBUILD to contain maintainer comment, got: %s", string(pkgbuildContent))
	}

	// Verify Git repository state
	checkGitRepo(t, pipelines.AUR.AURDir, "master")
}
