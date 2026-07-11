package templates_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"go-cli-package/internal/templates"
)

func TestEmbeddedTemplates(t *testing.T) {
	requiredFiles := []string{
		"config.mjs.tmpl",
		"sidebar.mjs.tmpl",
		"command.md.tmpl",
		"page.md.tmpl",
		"astro.config.mjs.tmpl",
		"custom.css.tmpl",
	}

	for _, name := range requiredFiles {
		t.Run(name, func(t *testing.T) {
			data, err := templates.FS.ReadFile(name)
			if err != nil {
				t.Fatalf("failed to read embedded file %s: %v", name, err)
			}
			if len(data) == 0 {
				t.Errorf("embedded file %s is empty", name)
			}
		})
	}
}

func TestRenderToFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "template_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tmpl, err := template.New("test").Parse("Hello {{ .Name }}!")
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	targetPath := filepath.Join(tempDir, "output.txt")
	data := struct {
		Name string
	}{
		Name: "World",
	}

	err = templates.RenderToFile(tmpl, targetPath, data)
	if err != nil {
		t.Fatalf("RenderToFile failed: %v", err)
	}

	content, err := ioutil.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read rendered file: %v", err)
	}

	expected := "Hello World!"
	if string(content) != expected {
		t.Errorf("expected %q, got %q", expected, string(content))
	}
}

func TestLoadFallsBackToEmbedded(t *testing.T) {
	tmpl, err := templates.Load("page.md.tmpl", []string{"/nonexistent/dir"})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if tmpl.Name() != "page.md.tmpl" {
		t.Errorf("expected embedded template name %q, got %q", "page.md.tmpl", tmpl.Name())
	}
}

func TestLoadOverrideDirectory(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "template_override_dir")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	overridePath := filepath.Join(tempDir, "page.md.tmpl")
	if err := os.WriteFile(overridePath, []byte("CUSTOM {{ .Title }}"), 0644); err != nil {
		t.Fatalf("failed to write override template: %v", err)
	}

	tmpl, err := templates.Load("page.md.tmpl", []string{tempDir})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	targetPath := filepath.Join(tempDir, "output.md")
	data := struct{ Title string }{Title: "Hi"}
	if err := templates.RenderToFile(tmpl, targetPath, data); err != nil {
		t.Fatalf("RenderToFile failed: %v", err)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read rendered file: %v", err)
	}
	if string(content) != "CUSTOM Hi" {
		t.Errorf("expected override content %q, got %q", "CUSTOM Hi", string(content))
	}
}

func TestLoadOverrideExactFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "template_override_file")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	overridePath := filepath.Join(tempDir, "page.md.tmpl")
	if err := os.WriteFile(overridePath, []byte("EXACT {{ .Title }}"), 0644); err != nil {
		t.Fatalf("failed to write override template: %v", err)
	}

	tmpl, err := templates.Load("page.md.tmpl", []string{overridePath})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	targetPath := filepath.Join(tempDir, "output.md")
	data := struct{ Title string }{Title: "Hi"}
	if err := templates.RenderToFile(tmpl, targetPath, data); err != nil {
		t.Fatalf("RenderToFile failed: %v", err)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read rendered file: %v", err)
	}
	if string(content) != "EXACT Hi" {
		t.Errorf("expected override content %q, got %q", "EXACT Hi", string(content))
	}
}
