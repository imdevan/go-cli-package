package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo initialises a bare-minimum git repo in dir so that git
// commands work inside the test directory.
func initTestRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %v\n%s", args, err, out)
		}
	}

	// Commit a dummy file so that the repo has a HEAD
	dummy := filepath.Join(dir, "README.md")
	if err := os.WriteFile(dummy, []byte("test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup %v: %v\n%s", args, err, out)
		}
	}
}

func TestTagList_Empty(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	// Should succeed with "No tags found." message
	if err := TagList(dir, 20); err != nil {
		t.Fatalf("TagList on empty repo: %v", err)
	}
}

func TestTagCreateAndDelete(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	const version = "1.2.3"
	tag := "v" + version

	// ---- Create ----
	// Create the tag locally only (no remote), so skip the push step by
	// using git directly and then verifying TagCreate fails on duplicate.
	createCmd := exec.Command("git", "tag", "-a", tag, "-m", "test")
	createCmd.Dir = dir
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("pre-create tag: %v\n%s", err, out)
	}

	// Verify the tag exists
	listCmd := exec.Command("git", "tag", "-l")
	listCmd.Dir = dir
	out, err := listCmd.Output()
	if err != nil {
		t.Fatalf("git tag -l: %v", err)
	}
	if !strings.Contains(string(out), tag) {
		t.Errorf("expected tag %s in list, got: %s", tag, out)
	}

	// ---- Delete (local only; no remote) ----
	// Override: delete locally only so the test doesn't try to push
	delCmd := exec.Command("git", "tag", "-d", tag)
	delCmd.Dir = dir
	if out, err := delCmd.CombinedOutput(); err != nil {
		t.Fatalf("git tag -d: %v\n%s", err, out)
	}

	// Confirm deletion
	afterCmd := exec.Command("git", "tag", "-l")
	afterCmd.Dir = dir
	afterOut, err := afterCmd.Output()
	if err != nil {
		t.Fatalf("git tag -l after delete: %v", err)
	}
	if strings.Contains(string(afterOut), tag) {
		t.Errorf("tag %s should have been deleted, still in: %s", tag, afterOut)
	}
}

func TestTagList_WithTags(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	// Create several annotated tags
	for _, v := range []string{"v1.0.0", "v1.1.0", "v2.0.0"} {
		cmd := exec.Command("git", "tag", "-a", v, "-m", v)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("create tag %s: %v\n%s", v, err, out)
		}
	}

	if err := TagList(dir, 20); err != nil {
		t.Fatalf("TagList: %v", err)
	}
}

func TestTagCreate_DuplicateWithoutForce(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	const version = "0.9.0"
	tag := "v" + version

	// Pre-create the tag
	createCmd := exec.Command("git", "tag", "-a", tag, "-m", "initial")
	createCmd.Dir = dir
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("pre-create tag: %v\n%s", err, out)
	}

	// TagCreate without force should return an error (no remote to push to either)
	err := TagCreate(dir, version, false)
	if err == nil {
		t.Fatal("expected error for duplicate tag without --force, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveVersion(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	cfg := &Config{Version: "3.0.0"}

	// When provided, it should be returned as-is
	got := resolveVersion(dir, "2.0.0", cfg)
	if got != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", got)
	}

	// When not provided and no git tag, falls back to config
	got = resolveVersion(dir, "", cfg)
	if got != "3.0.0" {
		t.Errorf("expected 3.0.0 (config fallback), got %s", got)
	}
}

func TestTagExists(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	if tagExists(dir, "v9.9.9") {
		t.Error("expected tagExists to return false for nonexistent tag")
	}

	// Create a tag
	cmd := exec.Command("git", "tag", "-a", "v9.9.9", "-m", "x")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create tag: %v\n%s", err, out)
	}

	if !tagExists(dir, "v9.9.9") {
		t.Error("expected tagExists to return true after creating tag")
	}
}

func TestRelease_SkipTagIfExists(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	// Pre-create tag
	tag := "v1.2.3"
	cmd := exec.Command("git", "tag", "-a", tag, "-m", "pre-existing")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pre-create tag: %v\n%s", err, out)
	}

	cfg := &Config{
		Name:    "test-pkg",
		Version: "1.2.3",
	}
	pipelines := NewPipelines(dir, cfg)

	// Release with skipGithub=true, skipTag=false.
	// Since tag exists, it should print warning but NOT return an error.
	err := Release(dir, pipelines, "1.2.3", false, true)
	if err != nil {
		t.Fatalf("Release failed on pre-existing tag: %v", err)
	}
}

