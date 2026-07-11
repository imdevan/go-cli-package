package main

import (
	"os"
	"testing"
)

func TestInitCmdNoOp(t *testing.T) {
	// Create a temporary docs directory to simulate "already exists".
	tmp := t.TempDir()
	docsPath := tmp + "/docs"
	if err := os.Mkdir(docsPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to the temp dir so runInit looks in the right place.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig) //nolint:errcheck
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// runInit should return nil (no-op) when docs/ already exists.
	if err := runInit("bunx", false); err != nil {
		t.Fatalf("expected no-op when docs/ exists, got: %v", err)
	}
}

func TestInitCmdMissingDocs(t *testing.T) {
	// Without docs/ present, runInit tries to shell out – just confirm the
	// command itself is wired up correctly (the cobra command returns non-nil).
	cmd := newInitCmd()
	if cmd == nil {
		t.Fatal("expected newInitCmd to return a non-nil command")
	}
	if cmd.Use != "init" {
		t.Fatalf("expected Use=init, got %q", cmd.Use)
	}
	// Verify --pkg-manager flag exists and defaults to bunx.
	f := cmd.Flags().Lookup("pkg-manager")
	if f == nil {
		t.Fatal("expected --pkg-manager flag")
	}
	if f.DefValue != "bun" {
		t.Fatalf("expected default bun, got %q", f.DefValue)
	}
}

func TestGenerateCmdWired(t *testing.T) {
	cmd := newGenerateCmd()
	if cmd == nil {
		t.Fatal("expected newGenerateCmd to return a non-nil command")
	}
	if cmd.Use != "generate" {
		t.Fatalf("expected Use=generate, got %q", cmd.Use)
	}
}

func TestWatchCmdWired(t *testing.T) {
	cmd := newWatchCmd()
	if cmd == nil {
		t.Fatal("expected newWatchCmd to return a non-nil command")
	}
	if cmd.Use != "watch" {
		t.Fatalf("expected Use=watch, got %q", cmd.Use)
	}
}

func TestRootHasSubcommands(t *testing.T) {
	root := newRootCmd()
	want := map[string]bool{
		"init":     false,
		"generate": false,
		"watch":    false,
	}
	for _, sub := range root.Commands() {
		if _, ok := want[sub.Use]; ok {
			want[sub.Use] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected subcommand %q to be registered on root", name)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	cases := []struct {
		path     string
		excluded bool
	}{
		{"node_modules/foo", true},
		{"docs/src/content/docs/bar.md", true},
		{".git/HEAD", true},
		{"cmd/go-cli-docs/root.go", false},
		{"internal/workflow/generate.go", false},
	}
	for _, tc := range cases {
		got := isExcluded(tc.path)
		if got != tc.excluded {
			t.Errorf("isExcluded(%q) = %v, want %v", tc.path, got, tc.excluded)
		}
	}
}

func TestWatchedExts(t *testing.T) {
	watched := []string{".md", ".go", ".toml"}
	for _, ext := range watched {
		if !watchedExts[ext] {
			t.Errorf("expected ext %q to be in watchedExts", ext)
		}
	}
	if watchedExts[".txt"] {
		t.Error("expected .txt to NOT be in watchedExts")
	}
}
