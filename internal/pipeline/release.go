package pipeline

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ziputil "github.com/imdevan/go-cli-package/internal/util/zip"
)

// ReleaseTarget describes a cross-compilation target.
type ReleaseTarget struct {
	GOOS   string
	GOARCH string
}

// DefaultReleaseTargets are the standard cross-platform targets.
var DefaultReleaseTargets = []ReleaseTarget{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
}

// GithubRelease builds cross-platform binaries, archives them, creates a
// GitHub release (via the `gh` CLI), and uploads the assets.
// It replaces github_release.sh.
func GithubRelease(rootDir string, cfg *Config, version string, cleanDist bool) error {
	version = strings.TrimPrefix(version, "v")
	tag := "v" + version
	name := cfg.Name
	cmdPkg := "./cmd/" + name
	distDir := filepath.Join(rootDir, "dist", tag)

	fmt.Printf("🔨 Building binaries for %s...\n", tag)

	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("create dist dir: %w", err)
	}

	var assets []string

	for _, t := range DefaultReleaseTargets {
		binaryName := fmt.Sprintf("%s-%s-%s", name, t.GOOS, t.GOARCH)
		if t.GOOS == "windows" {
			binaryName += ".exe"
		}
		binPath := filepath.Join(distDir, binaryName)

		fmt.Printf("  - %s/%s...\n", t.GOOS, t.GOARCH)

		buildCmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", binPath, cmdPkg)
		buildCmd.Dir = rootDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		buildCmd.Env = append(os.Environ(),
			"GOOS="+t.GOOS,
			"GOARCH="+t.GOARCH,
			"CGO_ENABLED=0",
		)
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("go build %s/%s: %w", t.GOOS, t.GOARCH, err)
		}

		// Create archive
		var archivePath string
		var archiveErr error

		if t.GOOS == "windows" {
			archivePath = filepath.Join(distDir, fmt.Sprintf("%s-%s-%s.zip", name, t.GOOS, t.GOARCH))
			archiveErr = createZip(binPath, archivePath, binaryName)
		} else {
			archivePath = filepath.Join(distDir, fmt.Sprintf("%s-%s-%s.tar.gz", name, t.GOOS, t.GOARCH))
			archiveErr = createTarGz(binPath, archivePath, binaryName)
		}
		if archiveErr != nil {
			return fmt.Errorf("create archive for %s/%s: %w", t.GOOS, t.GOARCH, archiveErr)
		}

		assets = append(assets, archivePath)
	}

	fmt.Printf("🚀 Creating GitHub release %s...\n", tag)

	// Verify gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found in PATH; install from https://cli.github.com")
	}

	cleanRepo := strings.TrimSuffix(cfg.Repository, "/")
	repoPath := strings.TrimPrefix(cleanRepo, "https://github.com/")

	ghArgs := []string{
		"release", "create", tag,
		"--title", tag,
		"--generate-notes",
		"--verify-tag",
	}
	if repoPath != "" && repoPath != cleanRepo {
		ghArgs = append(ghArgs, "-R", repoPath)
	}
	ghArgs = append(ghArgs, assets...)

	ghCmd := exec.Command("gh", ghArgs...)
	ghCmd.Dir = rootDir
	ghCmd.Stdout = os.Stdout
	ghCmd.Stderr = os.Stderr
	if err := ghCmd.Run(); err != nil {
		return fmt.Errorf("gh release create: %w", err)
	}

	fmt.Printf("✅ GitHub release %s created\n", tag)

	if cleanDist {
		fmt.Printf("🧹 Cleaning up dist directory %s...\n", distDir)
		if err := os.RemoveAll(distDir); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to clean up dist: %v\n", err)
		}
	}

	return nil
}

// Release is the end-to-end release command that:
//  1. Creates and pushes a git tag.
//  2. Builds cross-platform binaries and creates a GitHub release.
func Release(rootDir string, pipelines *Pipelines, version string, skipTag, skipGithub bool) error {
	cfg := pipelines.Config

	version = resolveVersion(rootDir, version, cfg)
	version = strings.TrimPrefix(version, "v")
	tag := "v" + version

	fmt.Printf("🚀 Starting release process for %s...\n", tag)
	fmt.Println()

	// Step 1 – Git tag
	if !skipTag {
		fmt.Println("Step 1: Creating git tag...")
		if tagExists(rootDir, tag) {
			fmt.Printf("⚠️  Tag %s already exists; skipping tag creation\n", tag)
			fmt.Printf("📤 Pushing existing tag %s to origin to ensure remote presence...\n", tag)
			if err := runCmdInDir(rootDir, "git", "push", "origin", tag); err != nil {
				fmt.Printf("⚠️  Warning: failed to push existing tag: %v\n", err)
			}
		} else {
			if err := TagCreate(rootDir, version, false); err != nil {
				return fmt.Errorf("tag: %w", err)
			}
		}
		fmt.Println()
	}

	// Step 2 – GitHub release
	if !skipGithub {
		fmt.Println("Step 2: Creating GitHub release...")
		if err := GithubRelease(rootDir, cfg, version, true); err != nil {
			return fmt.Errorf("github release: %w", err)
		}
		fmt.Println()
	}

	fmt.Printf("✅ Release %s complete!\n", tag)
	fmt.Println()
	fmt.Println("Next steps:")
	pkgName := cfg.GetPackageName()
	fmt.Printf("  1. Update package manifests: go-cli-package update all %s\n", version)
	fmt.Printf("  2. Test AUR: cd aur-%s && makepkg -si\n", pkgName)
	fmt.Printf("  3. Test Homebrew: brew install --build-from-source homebrew-%s/Formula/%s.rb\n", pkgName, pkgName)
	fmt.Printf("  4. Deploy: go-cli-package deploy all %s\n", version)
	return nil
}

// createTarGz creates a gzip-compressed tar archive at dst containing the
// single file at src, stored under the given name inside the archive.
func createTarGz(src, dst, name string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	hdr := &tar.Header{
		Name:    name,
		Mode:    int64(info.Mode()),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	_, err = io.Copy(tw, srcFile)
	return err
}

// createZip creates a zip archive at dst containing the single file at src
// stored under the given name. Uses a simple zip format without compression
// so that no additional libraries are needed.
func createZip(src, dst, name string) error {
	// Use the system zip command for simplicity and compatibility.
	// Alternatively, we could use archive/zip but exec keeps the binary small.
	zipPath, err := exec.LookPath("zip")
	if err != nil {
		// Fallback: use archive/zip from stdlib
		return createZipStdlib(src, dst, name)
	}

	cmd := exec.Command(zipPath, "-j", dst, src)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = name // zip uses the source filename
	return cmd.Run()
}

// createZipStdlib is a stdlib fallback for createZip when the zip binary is unavailable.
func createZipStdlib(src, dst, name string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	return ziputil.WriteEntry(dstFile, name, src)
}
