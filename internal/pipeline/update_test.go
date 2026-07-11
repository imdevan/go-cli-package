package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupInitializedPipelines(t *testing.T) (string, *Pipelines) {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := &Config{
		Name:        "test-cli",
		PackageName: "test-cli",
		Version:     "1.0.0",
		Homepage:    "https://example.com",
		Repository:  "https://github.com/org/test-cli",
		Description: "A test CLI",
		Author:      "test@example.com",
	}
	p := NewPipelines(tmpDir, cfg)
	if err := InitializeHomebrew(p.Homebrew, true); err != nil {
		t.Fatalf("InitializeHomebrew: %v", err)
	}
	if err := InitializeAUR(p.AUR, true); err != nil {
		t.Fatalf("InitializeAUR: %v", err)
	}
	return tmpDir, p
}

func TestUpdateHomebrew(t *testing.T) {
	_, p := setupInitializedPipelines(t)

	sha256s := map[string]string{
		"darwin-amd64": "aaa111",
		"darwin-arm64": "bbb222",
		"linux-amd64":  "ccc333",
		"linux-arm64":  "ddd444",
	}

	if err := UpdateHomebrew(p.Homebrew, "2.0.0", sha256s); err != nil {
		t.Fatalf("UpdateHomebrew: %v", err)
	}

	content, err := os.ReadFile(p.Homebrew.FormulaPath)
	if err != nil {
		t.Fatalf("read formula: %v", err)
	}
	s := string(content)

	for _, want := range []string{`version "2.0.0"`, "aaa111", "bbb222", "ccc333", "ddd444"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected formula to contain %q", want)
		}
	}
}

func TestUpdateAUR(t *testing.T) {
	_, p := setupInitializedPipelines(t)

	sha256s := map[string]string{
		"linux-amd64": "eee555",
		"linux-arm64": "fff666",
	}

	if err := UpdateAUR(p.AUR, "2.0.0", sha256s); err != nil {
		t.Fatalf("UpdateAUR: %v", err)
	}

	content, err := os.ReadFile(p.AUR.PKGBUILDPath)
	if err != nil {
		t.Fatalf("read PKGBUILD: %v", err)
	}
	s := string(content)

	for _, want := range []string{"pkgver=2.0.0", "eee555", "fff666"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected PKGBUILD to contain %q", want)
		}
	}
}

func TestParseSHA256FlagsInUpdatePublish(t *testing.T) {
	// parseSHA256Flags lives in cmd package; test the same logic here inline
	flags := []string{"linux-amd64=abc", "darwin-arm64=xyz"}
	m := make(map[string]string)
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	if m["linux-amd64"] != "abc" || m["darwin-arm64"] != "xyz" {
		t.Errorf("unexpected map: %v", m)
	}
}

func TestUpdateHomebrewMissingDir(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{Name: "missing", PackageName: "missing", Version: "1.0.0"}
	p := NewPipelines(tmpDir, cfg)

	// Don't init — expect an error
	err := UpdateHomebrew(p.Homebrew, "1.0.0", map[string]string{})
	if err == nil {
		t.Fatal("expected error when tap directory does not exist")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}

	// Also clean up any accidental directory
	_ = os.RemoveAll(filepath.Join(tmpDir, "homebrew-missing"))
}
