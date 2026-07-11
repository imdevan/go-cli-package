package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imdevan/go-cli-package/internal/testutil"
	"github.com/spf13/cobra"
)

func runCLIWithStdoutCapture(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	os.Stderr = w

	cmd.SetOut(w)
	cmd.SetErr(w)
	cmd.SetArgs(args)

	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd.Execute()
	}()

	var buf bytes.Buffer
	doneChan := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(doneChan)
	}()

	cmdErr := <-errChan
	w.Close()
	<-doneChan

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return buf.String(), cmdErr
}

func TestCLICommands(t *testing.T) {
	// Remember original working directory to restore it later
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	tmpDir := testutil.WithTempWorkspace(t)

	// Create internal/package/package.toml in tmpDir
	pkgTomlDir := filepath.Join(tmpDir, "internal", "package")
	if err := os.MkdirAll(pkgTomlDir, 0755); err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	
	packageTomlContent := `
name = "cli-test"
package_name = "cli-test"
module = "cli-test"
description = "A cli test"
version = "1.0.0"
homepage = "https://example.com"
repository = "https://github.com/user/cli-test"
author = "test@example.com"
`
	if err := os.WriteFile(filepath.Join(pkgTomlDir, "package.toml"), []byte(packageTomlContent), 0644); err != nil {
		t.Fatalf("failed to write package.toml: %v", err)
	}

	rootCmd := newRootCmd()

	// Test "init" command
	out, err := runCLIWithStdoutCapture(t, rootCmd, "init", "all", "--force")
	if err != nil {
		t.Fatalf("init all command failed: %v, output: %s", err, out)
	}

	if !strings.Contains(out, "Homebrew tap initialized") {
		t.Errorf("expected output to contain Homebrew tap initialized, got: %s", out)
	}
	if !strings.Contains(out, "AUR repository initialized") {
		t.Errorf("expected output to contain AUR repository initialized, got: %s", out)
	}

	// Verify initialized files
	if _, err := os.Stat(filepath.Join(tmpDir, "homebrew-cli-test", "Formula", "cli-test.rb")); err != nil {
		t.Errorf("expected Homebrew formula to be created, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "aur-cli-test", "PKGBUILD")); err != nil {
		t.Errorf("expected AUR PKGBUILD to be created, got: %v", err)
	}

	// Test "build" command
	// For build, we need a template file for AUR under <rootDir>/aur/PKGBUILD.template
	aurDir := filepath.Join(tmpDir, "aur")
	if err := os.MkdirAll(aurDir, 0755); err != nil {
		t.Fatalf("failed to create aur dir: %v", err)
	}
	templateContent := `pkgname=__PKGNAME__
pkgver=__PKGVER__
source=("${pkgname}-${pkgver}.tar.gz::__SOURCE_URL__")
sha256sums=('__SHA256__')
`
	if err := os.WriteFile(filepath.Join(aurDir, "PKGBUILD.template"), []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write PKGBUILD template: %v", err)
	}

	// We also need cmd/cli-test/main.go to test go build
	cmdPkgDir := filepath.Join(tmpDir, "cmd", "cli-test")
	if err := os.MkdirAll(cmdPkgDir, 0755); err != nil {
		t.Fatalf("failed to create cmd pkg dir: %v", err)
	}
	mainGoContent := `package main
func main() {}
`
	if err := os.WriteFile(filepath.Join(cmdPkgDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Write a mock go.mod so go build works inside tmpDir
	goModContent := `module cli-test
go 1.25.6
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Run "build" command
	out, err = runCLIWithStdoutCapture(t, rootCmd, "build", "all", "--sha256", "abc123sha256")
	if err != nil {
		t.Fatalf("build command failed: %v, output: %s", err, out)
	}

	if !strings.Contains(out, "Building Go binary") {
		t.Errorf("expected output to contain Building Go binary, got: %s", out)
	}
	if !strings.Contains(out, "Generating AUR PKGBUILD") {
		t.Errorf("expected output to contain Generating AUR PKGBUILD, got: %s", out)
	}

	// Verify build outputs
	if _, err := os.Stat(filepath.Join(tmpDir, "bin", "cli-test")); err != nil {
		t.Errorf("expected built binary to exist, got: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "dist", "aur", "PKGBUILD")); err != nil {
		t.Errorf("expected generated PKGBUILD to exist, got: %v", err)
	}
}
