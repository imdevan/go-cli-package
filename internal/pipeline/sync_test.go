package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateReadmeH1(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		newName  string
		expected string
	}{
		{
			name: "basic H1 update",
			input: `# old-name

Some description.`,
			newName: "new-name",
			expected: `# new-name

Some description.`,
		},
		{
			name: "ignore code block comment",
			input: `# old-name

` + "```bash" + `
# Comment in code block
` + "```" + `
Some description.`,
			newName: "new-name",
			expected: `# new-name

` + "```bash" + `
# Comment in code block
` + "```" + `
Some description.`,
		},
		{
			name: "update first H1 only",
			input: `# first-h1

# second-h1`,
			newName: "updated-name",
			expected: `# updated-name

# second-h1`,
		},
		{
			name:     "no H1 present",
			input:    "Only some text without any headers.",
			newName:  "new-name",
			expected: "Only some text without any headers.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			readmePath := filepath.Join(tmpDir, "README.md")
			if err := os.WriteFile(readmePath, []byte(tt.input), 0644); err != nil {
				t.Fatalf("failed to write test README: %v", err)
			}

			err := updateReadmeH1(tmpDir, tt.newName)
			if err != nil {
				t.Fatalf("updateReadmeH1 returned error: %v", err)
			}

			content, err := os.ReadFile(readmePath)
			if err != nil {
				t.Fatalf("failed to read updated README: %v", err)
			}

			got := string(content)
			if got != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}

func TestSyncReadmeH1(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create mock structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "internal", "package"), 0755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	// Mock package.toml
	tomlContent := `
name = "new-cool-package"
module = "github.com/user/new-cool-package"
description = "A very cool package description."
version = "1.0.0"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "internal", "package", "package.toml"), []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write package.toml: %v", err)
	}

	// Mock go.mod
	goModContent := "module github.com/user/old-package\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Mock justfile
	justfileContent := "PACKAGE := \"old-package\"\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "justfile"), []byte(justfileContent), 0644); err != nil {
		t.Fatalf("failed to write justfile: %v", err)
	}

	// Mock .github/workflows files
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("failed to create workflows dir: %v", err)
	}
	releaseYML := "name: release\nrun: go build -o bin/old-package ./cmd/old-package\n"
	if err := os.WriteFile(filepath.Join(workflowsDir, "release.yml"), []byte(releaseYML), 0644); err != nil {
		t.Fatalf("failed to write release.yml: %v", err)
	}

	// Mock README.md
	readmeContent := `# old-package

This is the old package description.`
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}

	// Run Sync
	err := Sync(tmpDir, "")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Check H1 and Description in README.md
	updatedReadme, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	if err != nil {
		t.Fatalf("failed to read updated README.md: %v", err)
	}

	readmeStr := string(updatedReadme)
	if !strings.Contains(readmeStr, "# new-cool-package") {
		t.Errorf("expected README to contain '# new-cool-package', got:\n%s", readmeStr)
	}
	if !strings.Contains(readmeStr, "A very cool package description.") {
		t.Errorf("expected README to contain 'A very cool package description.', got:\n%s", readmeStr)
	}

	// Check .github/workflows/release.yml was updated
	updatedWorkflow, err := os.ReadFile(filepath.Join(tmpDir, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("failed to read updated release.yml: %v", err)
	}
	workflowStr := string(updatedWorkflow)
	if strings.Contains(workflowStr, "old-package") {
		t.Errorf("expected release.yml to have no references to 'old-package', got:\n%s", workflowStr)
	}
	if !strings.Contains(workflowStr, "new-cool-package") {
		t.Errorf("expected release.yml to contain 'new-cool-package', got:\n%s", workflowStr)
	}
}

func TestUpdateGithubWorkflows(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string // relative path -> content
		oldName  string
		newName  string
		expected map[string]string // relative path -> expected content
	}{
		{
			name: "updates name in yml files",
			files: map[string]string{
				"release.yml": "run: go build -o bin/myapp ./cmd/myapp\n",
				"docs.yml":    "name: Deploy docs for myapp\n",
			},
			oldName: "myapp",
			newName: "newapp",
			expected: map[string]string{
				"release.yml": "run: go build -o bin/newapp ./cmd/newapp\n",
				"docs.yml":    "name: Deploy docs for newapp\n",
			},
		},
		{
			name: "skips non-yml files",
			files: map[string]string{
				"release.yml": "run: go build -o bin/myapp ./cmd/myapp\n",
				"notes.txt":   "binary is myapp\n",
			},
			oldName: "myapp",
			newName: "newapp",
			expected: map[string]string{
				"release.yml": "run: go build -o bin/newapp ./cmd/newapp\n",
				"notes.txt":   "binary is myapp\n", // unchanged
			},
		},
		{
			name: "no-op when workflows dir absent",
			files: map[string]string{},
			oldName: "myapp",
			newName: "newapp",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			workflowsDir := filepath.Join(tmpDir, ".github", "workflows")

			if len(tt.files) > 0 {
				if err := os.MkdirAll(workflowsDir, 0755); err != nil {
					t.Fatalf("failed to create workflows dir: %v", err)
				}
				for name, content := range tt.files {
					if err := os.WriteFile(filepath.Join(workflowsDir, name), []byte(content), 0644); err != nil {
						t.Fatalf("failed to write %s: %v", name, err)
					}
				}
			}

			if err := updateGithubWorkflows(tmpDir, tt.oldName, tt.newName); err != nil {
				t.Fatalf("updateGithubWorkflows returned error: %v", err)
			}

			for name, want := range tt.expected {
				got, err := os.ReadFile(filepath.Join(workflowsDir, name))
				if err != nil {
					t.Fatalf("failed to read %s: %v", name, err)
				}
				if string(got) != want {
					t.Errorf("%s: expected %q, got %q", name, want, string(got))
				}
			}
		})
	}
}
