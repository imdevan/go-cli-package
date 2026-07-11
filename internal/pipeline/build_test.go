package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAUR(t *testing.T) {
	tmpDir := t.TempDir()

	// Create aur directory and PKGBUILD.template
	aurDir := filepath.Join(tmpDir, "aur")
	if err := os.MkdirAll(aurDir, 0755); err != nil {
		t.Fatalf("failed to create aur dir: %v", err)
	}

	templateContent := `pkgname=__PKGNAME__
pkgver=__PKGVER__
pkgdesc="__DESCRIPTION__"
url="__HOMEPAGE__"
source=("${pkgname}-${pkgver}.tar.gz::__SOURCE_URL__")
sha256sums=('__SHA256__')
`
	templatePath := filepath.Join(aurDir, "PKGBUILD.template")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	cfg := &Config{
		Name:        "my-test-app",
		PackageName: "my-test-app-package",
		Version:     "1.0.0",
		Homepage:    "https://example.com/test",
		Repository:  "https://github.com/user/my-test-app",
		Description: "A fine description",
	}

	err := BuildAUR(tmpDir, cfg, "v1.2.3", "my-sha256-hash-value")
	if err != nil {
		t.Fatalf("BuildAUR failed: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "dist", "aur", "PKGBUILD")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected PKGBUILD to exist at %s: %v", outputPath, err)
	}

	contentBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output PKGBUILD: %v", err)
	}

	content := string(contentBytes)
	if !strings.Contains(content, "pkgname=my-test-app-package") {
		t.Errorf("expected pkgname replacement, got: %s", content)
	}
	if !strings.Contains(content, "pkgver=1.2.3") {
		t.Errorf("expected pkgver replacement, got: %s", content)
	}
	if !strings.Contains(content, "pkgdesc=\"A fine description\"") {
		t.Errorf("expected pkgdesc replacement, got: %s", content)
	}
	if !strings.Contains(content, "url=\"https://example.com/test\"") {
		t.Errorf("expected url replacement, got: %s", content)
	}
	if !strings.Contains(content, `source=("${pkgname}-${pkgver}.tar.gz::https://github.com/user/my-test-app/archive/refs/tags/v1.2.3.tar.gz")`) {
		t.Errorf("expected source replacement, got: %s", content)
	}
	if !strings.Contains(content, "sha256sums=('my-sha256-hash-value')") {
		t.Errorf("expected sha256sums replacement, got: %s", content)
	}
}

func TestBuildBinary(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	repoRoot := filepath.Dir(filepath.Dir(cwd))

	// Create a temporary package under the real cmd/ directory
	testCmdDir := filepath.Join(repoRoot, "cmd", "test-binary-pkg")
	if err := os.MkdirAll(testCmdDir, 0755); err != nil {
		t.Fatalf("failed to create cmd/test-binary-pkg: %v", err)
	}
	defer os.RemoveAll(testCmdDir)

	mainCode := `package main
import "fmt"
func main() {
	fmt.Println("hello test")
}
`
	if err := os.WriteFile(filepath.Join(testCmdDir, "main.go"), []byte(mainCode), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	cfg := &Config{
		Name: "test-binary-pkg",
	}

	// Build the binary
	err = BuildBinary(repoRoot, cfg)
	if err != nil {
		t.Fatalf("BuildBinary failed: %v", err)
	}

	binPath := filepath.Join(repoRoot, "bin", "test-binary-pkg")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("expected built binary to exist at %s: %v", binPath, err)
	}
	defer os.Remove(binPath)
}
