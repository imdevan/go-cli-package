package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Sync synchronizes the project files (go.mod, imports, justfile, cmd dir, README)
// with the values specified in the given package.toml.
func Sync(rootDir string, configPath string) error {
	var cfg *Config
	var err error

	if configPath != "" {
		cfg, err = LoadConfig(configPath)
	} else {
		cfg, err = FindAndLoadConfig(rootDir)
	}
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Name == "" {
		return fmt.Errorf("'name' is required in config")
	}
	if cfg.Module == "" {
		return fmt.Errorf("'module' is required in config")
	}

	fmt.Printf("Syncing project from package.toml...\n")
	fmt.Printf("  Project Name: %s\n", cfg.Name)
	fmt.Printf("  Module Name:  %s\n", cfg.Module)
	fmt.Printf("  Description:  %s\n", cfg.Description)
	fmt.Printf("  Short:        %s\n", cfg.Short)
	fmt.Printf("  Version:      %s\n", cfg.Version)
	fmt.Println()

	currentModule, err := getCurrentModule(rootDir)
	if err != nil {
		return fmt.Errorf("failed to get current module from go.mod: %w", err)
	}

	currentName, err := getCurrentName(rootDir)
	if err != nil {
		return fmt.Errorf("failed to get current project name: %w", err)
	}

	fmt.Println("Syncing files...")

	// Update go.mod
	if currentModule != cfg.Module {
		fmt.Println("Updating go.mod module name...")
		if err := updateGoMod(rootDir, currentModule, cfg.Module); err != nil {
			return fmt.Errorf("failed to update go.mod: %w", err)
		}

		// Update all Go import paths
		fmt.Println("Updating Go import paths...")
		if err := updateGoImports(rootDir, currentModule, cfg.Module); err != nil {
			return fmt.Errorf("failed to update Go import paths: %w", err)
		}
	}

	// Update config paths
	if currentName != cfg.Name {
		// Rename cmd directory first
		oldCmdDir := filepath.Join(rootDir, "cmd", currentName)
		newCmdDir := filepath.Join(rootDir, "cmd", cfg.Name)
		if _, err := os.Stat(oldCmdDir); err == nil {
			fmt.Printf("Renaming %s to %s...\n", filepath.Join("cmd", currentName), filepath.Join("cmd", cfg.Name))
			if err := os.Rename(oldCmdDir, newCmdDir); err != nil {
				return fmt.Errorf("failed to rename cmd directory: %w", err)
			}
		}

		// Update completion examples (after directory rename)
		completionFile := filepath.Join(rootDir, "cmd", cfg.Name, "completion.go")
		if _, err := os.Stat(completionFile); err == nil {
			fmt.Println("Updating completion examples...")
			if err := updateCompletionFile(completionFile, currentName, cfg.Name); err != nil {
				return fmt.Errorf("failed to update completion file: %w", err)
			}
		}

		// Update justfile
		fmt.Println("Updating justfile...")
		if err := updateJustfile(rootDir, currentName, cfg.Name); err != nil {
			return fmt.Errorf("failed to update justfile: %w", err)
		}
	}

	// Update README description
	if cfg.Description != "" {
		fmt.Println("Updating README description...")
		if err := updateReadmeDescription(rootDir, cfg.Description); err != nil {
			return fmt.Errorf("failed to update README.md: %w", err)
		}
	}

	fmt.Println("✓ Sync complete!")
	return nil
}

func getCurrentModule(rootDir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(rootDir, "go.mod"))
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(line[7:]), nil
		}
	}
	return "", fmt.Errorf("module declaration not found in go.mod")
}

var justfilePackageRegexp = regexp.MustCompile(`PACKAGE\s*:=\s*"([^"]+)"`)

func getCurrentName(rootDir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(rootDir, "justfile"))
	if err == nil {
		matches := justfilePackageRegexp.FindStringSubmatch(string(content))
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Fallback to checking cmd directory
	files, err := os.ReadDir(filepath.Join(rootDir, "cmd"))
	if err == nil {
		for _, f := range files {
			if f.IsDir() && f.Name() != "completion" {
				return f.Name(), nil
			}
		}
	}
	return "", fmt.Errorf("current project name not found in justfile or cmd/")
}

func updateGoMod(rootDir, currentModule, newModule string) error {
	goModPath := filepath.Join(rootDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	oldLine := "module " + currentModule
	newLine := "module " + newModule
	updated := strings.Replace(string(content), oldLine, newLine, 1)
	return os.WriteFile(goModPath, []byte(updated), 0644)
}

func updateGoImports(rootDir, currentModule, newModule string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "dist" || info.Name() == "bin" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".go" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			oldImport := "\"" + currentModule + "/"
			newImport := "\"" + newModule + "/"
			oldImportRaw := "`" + currentModule + "/"
			newImportRaw := "`" + newModule + "/"

			contentStr := string(content)
			updated := false
			if strings.Contains(contentStr, oldImport) {
				contentStr = strings.ReplaceAll(contentStr, oldImport, newImport)
				updated = true
			}
			if strings.Contains(contentStr, oldImportRaw) {
				contentStr = strings.ReplaceAll(contentStr, oldImportRaw, newImportRaw)
				updated = true
			}

			if updated {
				err = os.WriteFile(path, []byte(contentStr), info.Mode())
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func updateCompletionFile(path, oldName, newName string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated := strings.ReplaceAll(string(content), oldName, newName)
	return os.WriteFile(path, []byte(updated), 0644)
}

func updateJustfile(rootDir, oldName, newName string) error {
	justfilePath := filepath.Join(rootDir, "justfile")
	content, err := os.ReadFile(justfilePath)
	if err != nil {
		return err
	}
	contentStr := string(content)
	
	oldLine := fmt.Sprintf(`PACKAGE := "%s"`, oldName)
	newLine := fmt.Sprintf(`PACKAGE := "%s"`, newName)
	if strings.Contains(contentStr, oldLine) {
		contentStr = strings.Replace(contentStr, oldLine, newLine, 1)
	} else {
		contentStr = strings.ReplaceAll(contentStr, "bin/"+oldName, "bin/"+newName)
		contentStr = strings.ReplaceAll(contentStr, "./cmd/"+oldName, "./cmd/"+newName)
	}
	return os.WriteFile(justfilePath, []byte(contentStr), 0644)
}

func updateReadmeDescription(rootDir, description string) error {
	readmePath := filepath.Join(rootDir, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	updated := false
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !updated && trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "<img") {
			newLines = append(newLines, description)
			updated = true
			continue
		}
		newLines = append(newLines, line)
	}

	return os.WriteFile(readmePath, []byte(strings.Join(newLines, "\n")), 0644)
}
