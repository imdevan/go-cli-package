package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureIgnored_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	err := ensureIgnored(tmpDir, "homebrew-foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	expected := "homebrew-foo/\n"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestEnsureIgnored_Append(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	initialContent := "foo\nbar"
	if err := os.WriteFile(gitignorePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write initial .gitignore: %v", err)
	}

	err := ensureIgnored(tmpDir, "aur-foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	expected := "foo\nbar\naur-foo/\n"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestEnsureIgnored_AlreadyIgnoredExact(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	initialContent := "foo\naur-foo/\nbar"
	if err := os.WriteFile(gitignorePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write initial .gitignore: %v", err)
	}

	err := ensureIgnored(tmpDir, "aur-foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if string(data) != initialContent {
		t.Errorf("expected no changes, got %q", string(data))
	}
}

func TestEnsureIgnored_AlreadyIgnoredWildcard(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	initialContent := "foo\nhomebrew-*/\nbar"
	if err := os.WriteFile(gitignorePath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write initial .gitignore: %v", err)
	}

	err := ensureIgnored(tmpDir, "homebrew-foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if string(data) != initialContent {
		t.Errorf("expected no changes, got %q", string(data))
	}
}
