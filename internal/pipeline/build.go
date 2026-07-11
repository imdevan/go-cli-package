package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildBinary builds the local Go binary for the package.
// It compiles the code in ./cmd/<name> and saves the output to ./bin/<name>.
func BuildBinary(rootDir string, cfg *Config) error {
	pkgName := cfg.Name
	binDir := filepath.Join(rootDir, "bin")
	binPath := filepath.Join(binDir, pkgName)

	fmt.Printf("🏗️  Building Go binary for %s...\n", pkgName)

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Run go build in rootDir
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/"+pkgName)
	cmd.Dir = rootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	// Get file size
	fi, err := os.Stat(binPath)
	if err != nil {
		return fmt.Errorf("failed to check built binary size: %w", err)
	}

	sizeMB := float64(fi.Size()) / (1024 * 1024)
	fmt.Printf("✅ Build complete! Binary size: %.2f MB\n", sizeMB)
	return nil
}

// BuildAUR generates the PKGBUILD file from the template.
func BuildAUR(rootDir string, cfg *Config, version string, sha256 string) error {
	pkgName := cfg.GetPackageName()
	templatePath := filepath.Join(rootDir, "aur", "PKGBUILD.template")
	outputDir := filepath.Join(rootDir, "dist", "aur")
	outputPath := filepath.Join(outputDir, "PKGBUILD")

	fmt.Printf("📦 Generating AUR PKGBUILD for %s...\n", pkgName)

	// Determine version
	if version == "" {
		version = getGitTagVersion(rootDir)
		if version == "" {
			version = cfg.Version
		}
	}
	// Trim 'v' prefix for PKGBUILD pkgver
	pkgVer := strings.TrimPrefix(version, "v")

	// Determine source URL
	cleanRepo := strings.TrimSuffix(cfg.Repository, "/")
	sourceURL := fmt.Sprintf("%s/archive/refs/tags/v%s.tar.gz", cleanRepo, pkgVer)

	if sha256 == "" {
		sha256 = os.Getenv("AUR_SOURCE_SHA256")
		if sha256 == "" {
			return fmt.Errorf("sha256 hash (or AUR_SOURCE_SHA256 env variable) is required to generate %s", outputPath)
		}
	}

	// Read template
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read AUR PKGBUILD template at %s: %w", templatePath, err)
	}

	content := string(templateBytes)
	content = strings.ReplaceAll(content, "__PKGNAME__", pkgName)
	content = strings.ReplaceAll(content, "__PKGVER__", pkgVer)
	content = strings.ReplaceAll(content, "__DESCRIPTION__", cfg.Description)
	content = strings.ReplaceAll(content, "__HOMEPAGE__", cfg.Homepage)
	content = strings.ReplaceAll(content, "__SOURCE_URL__", sourceURL)
	content = strings.ReplaceAll(content, "__SHA256__", sha256)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write PKGBUILD to %s: %w", outputPath, err)
	}

	fmt.Printf("✅ Wrote %s\n", outputPath)
	return nil
}

func getGitTagVersion(rootDir string) string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = rootDir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
