package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveVersion_Provided(t *testing.T) {
	got, err := ResolveVersion("1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", got)
	}
}

func TestResolveVersion_FromPackageToml(t *testing.T) {
	// Remember original working directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}

	// Create internal/package/package.toml in tmpDir
	pkgDir := filepath.Join(tmpDir, "internal", "package")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}

	packageTomlContent := `
name = "test-pkg"
version = "4.5.6"
`
	if err := os.WriteFile(filepath.Join(pkgDir, "package.toml"), []byte(packageTomlContent), 0644); err != nil {
		t.Fatalf("failed to write package.toml: %v", err)
	}

	got, err := ResolveVersion("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "4.5.6" {
		t.Errorf("expected 4.5.6, got %s", got)
	}
}

func TestResolveVersion_Missing(t *testing.T) {
	// Remember original working directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}

	_, err = ResolveVersion("")
	if err == nil {
		t.Fatal("expected error when package.toml is missing, got nil")
	}
}
