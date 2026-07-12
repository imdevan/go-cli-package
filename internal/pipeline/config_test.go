package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	tomlPath := filepath.Join(tmpDir, "package.toml")

	content := `
name = "my-test-app"
package_name = "my-test-app-package"
homebrew_package_name = "my-test-app-brew"
aur_package_name = "my-test-app-aur"
module = "github.com/user/my-test-app"
description = "A test application"
short = "Test app"
version = "1.2.3"
homepage = "https://example.com"
repository = "https://github.com/user/my-test-app"
author = "test@example.com"
docs_site = "https://example.com/docs"
docs_base = "/docs"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test toml: %v", err)
	}

	cfg, err := LoadConfig(tomlPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Name != "my-test-app" {
		t.Errorf("expected Name to be 'my-test-app', got '%s'", cfg.Name)
	}
	if cfg.GetPackageName() != "my-test-app-package" {
		t.Errorf("expected GetPackageName to be 'my-test-app-package', got '%s'", cfg.GetPackageName())
	}
	if cfg.GetHomebrewPackageName() != "my-test-app-brew" {
		t.Errorf("expected GetHomebrewPackageName to be 'my-test-app-brew', got '%s'", cfg.GetHomebrewPackageName())
	}
	if cfg.GetAURPackageName() != "my-test-app-aur" {
		t.Errorf("expected GetAURPackageName to be 'my-test-app-aur', got '%s'", cfg.GetAURPackageName())
	}
	if cfg.Version != "1.2.3" {
		t.Errorf("expected Version to be '1.2.3', got '%s'", cfg.Version)
	}
}

func TestFindAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	pkgTomlDir := filepath.Join(tmpDir, "internal", "package")
	if err := os.MkdirAll(pkgTomlDir, 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}
	tomlPath := filepath.Join(pkgTomlDir, "package.toml")
	content := `
name = "test-pkg"
version = "0.0.1"
`
	if err := os.WriteFile(tomlPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test toml: %v", err)
	}

	cfg, err := FindAndLoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("FindAndLoadConfig failed: %v", err)
	}
	if cfg.Name != "test-pkg" {
		t.Errorf("expected Name to be 'test-pkg', got '%s'", cfg.Name)
	}
	if cfg.GetHomebrewPackageName() != "test-pkg" {
		t.Errorf("expected GetHomebrewPackageName to default to Name, got '%s'", cfg.GetHomebrewPackageName())
	}
	if cfg.GetAURPackageName() != "test-pkg" {
		t.Errorf("expected GetAURPackageName to default to Name, got '%s'", cfg.GetAURPackageName())
	}
}

func TestNewPipelines(t *testing.T) {
	cfg := &Config{
		Name:        "awesome-cli",
		PackageName: "awesome-cli",
		Version:     "2.0.0",
		Homepage:    "https://example.com",
		Repository:  "https://github.com/org/awesome-cli",
		Description: "An awesome command line interface",
	}

	pipelines := NewPipelines("/tmp/workspace", cfg)

	if pipelines.Homebrew.TapName != "homebrew-awesome-cli" {
		t.Errorf("expected TapName to be 'homebrew-awesome-cli', got '%s'", pipelines.Homebrew.TapName)
	}
	if pipelines.Homebrew.TapDir != "/tmp/workspace/homebrew-awesome-cli" {
		t.Errorf("expected TapDir to be '/tmp/workspace/homebrew-awesome-cli', got '%s'", pipelines.Homebrew.TapDir)
	}
	if pipelines.Homebrew.FormulaPath != "/tmp/workspace/homebrew-awesome-cli/Formula/awesome-cli.rb" {
		t.Errorf("expected FormulaPath to be '/tmp/workspace/homebrew-awesome-cli/Formula/awesome-cli.rb', got '%s'", pipelines.Homebrew.FormulaPath)
	}
	if pipelines.Homebrew.ClassName != "AwesomeCli" {
		t.Errorf("expected ClassName to be 'AwesomeCli', got '%s'", pipelines.Homebrew.ClassName)
	}

	if pipelines.AUR.AURDir != "/tmp/workspace/aur-awesome-cli" {
		t.Errorf("expected AURDir to be '/tmp/workspace/aur-awesome-cli', got '%s'", pipelines.AUR.AURDir)
	}
	if pipelines.AUR.PKGBUILDPath != "/tmp/workspace/aur-awesome-cli/PKGBUILD" {
		t.Errorf("expected PKGBUILDPath to be '/tmp/workspace/aur-awesome-cli/PKGBUILD', got '%s'", pipelines.AUR.PKGBUILDPath)
	}
}

func TestNewPipelinesOverrides(t *testing.T) {
	cfg := &Config{
		Name:                "awesome-cli",
		PackageName:         "awesome-cli-pkg",
		HomebrewPackageName: "awesome-homebrew",
		AURPackageName:      "awesome-aur",
		Version:             "2.0.0",
		Homepage:            "https://example.com",
		Repository:          "https://github.com/org/awesome-cli",
		Description:         "An awesome command line interface",
	}

	pipelines := NewPipelines("/tmp/workspace", cfg)

	if pipelines.Homebrew.TapName != "homebrew-awesome-homebrew" {
		t.Errorf("expected TapName to be 'homebrew-awesome-homebrew', got '%s'", pipelines.Homebrew.TapName)
	}
	if pipelines.Homebrew.TapDir != "/tmp/workspace/homebrew-awesome-homebrew" {
		t.Errorf("expected TapDir to be '/tmp/workspace/homebrew-awesome-homebrew', got '%s'", pipelines.Homebrew.TapDir)
	}
	if pipelines.Homebrew.FormulaPath != "/tmp/workspace/homebrew-awesome-homebrew/Formula/awesome-homebrew.rb" {
		t.Errorf("expected FormulaPath to be '/tmp/workspace/homebrew-awesome-homebrew/Formula/awesome-homebrew.rb', got '%s'", pipelines.Homebrew.FormulaPath)
	}
	if pipelines.Homebrew.ClassName != "AwesomeHomebrew" {
		t.Errorf("expected ClassName to be 'AwesomeHomebrew', got '%s'", pipelines.Homebrew.ClassName)
	}

	if pipelines.AUR.AURDir != "/tmp/workspace/aur-awesome-aur" {
		t.Errorf("expected AURDir to be '/tmp/workspace/aur-awesome-aur', got '%s'", pipelines.AUR.AURDir)
	}
	if pipelines.AUR.PKGBUILDPath != "/tmp/workspace/aur-awesome-aur/PKGBUILD" {
		t.Errorf("expected PKGBUILDPath to be '/tmp/workspace/aur-awesome-aur/PKGBUILD', got '%s'", pipelines.AUR.PKGBUILDPath)
	}
}
