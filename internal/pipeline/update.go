package pipeline

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// downloadAndHash fetches url to a temp file and returns its hex SHA256.
func downloadAndHash(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: status %s", url, resp.Status)
	}

	tmp, err := os.CreateTemp("", "go-cli-package-sha256-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), resp.Body); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// localSHA256 returns the hex SHA256 of a local file.
func localSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// resolveReleaseSHA resolves the SHA256 for a given platform release asset.
// It first looks in dist/v<version>/ for a locally built asset, then downloads.
func resolveReleaseSHA(rootDir, name, version, platform, ext string) (string, error) {
	local := filepath.Join(rootDir, "dist", "v"+version, name+"-"+platform+"."+ext)
	if _, err := os.Stat(local); err == nil {
		fmt.Fprintf(os.Stderr, "📎 Using local asset: %s\n", local)
		return localSHA256(local)
	}

	cleanRepo := "" // caller must pass URL directly; here just signal to caller
	_ = cleanRepo
	return "", fmt.Errorf("local asset not found at %s; provide --sha256 or pre-build release assets", local)
}

// UpdateHomebrew rewrites the Homebrew formula with a new version and SHAs.
// If sha256s map is nil/empty, it tries to download assets from GitHub releases.
func UpdateHomebrew(p *HomebrewPipeline, version string, sha256s map[string]string) error {
	if version == "" {
		version = getGitTagVersion(p.Config.Repository)
		if version == "" {
			version = p.Config.Version
		}
	}
	version = strings.TrimPrefix(version, "v")

	if _, err := os.Stat(p.TapDir); os.IsNotExist(err) {
		return fmt.Errorf("homebrew tap not found at %s; run 'init homebrew' first", p.TapDir)
	}

	cleanRepo := strings.TrimSuffix(p.Config.Repository, "/")
	repoPath := strings.TrimPrefix(cleanRepo, "https://github.com/")

	platforms := []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64"}

	// Resolve SHAs, downloading if not provided
	resolved := make(map[string]string)
	for _, plat := range platforms {
		if sha256s != nil {
			if v, ok := sha256s[plat]; ok && v != "" {
				resolved[plat] = v
				continue
			}
		}
		url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s-%s.tar.gz",
			repoPath, version, p.Config.Name, plat)
		fmt.Fprintf(os.Stderr, "📥 Downloading to calculate SHA256 for %s...\n", plat)
		sha, err := downloadAndHash(url)
		if err != nil {
			return fmt.Errorf("failed to resolve SHA256 for %s: %w", plat, err)
		}
		resolved[plat] = sha
	}

	className := p.ClassName
	formulaTmpl := `class {{.ClassName}} < Formula
  desc "{{.Description}}"
  homepage "{{.Homepage}}"
  version "{{.Version}}"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/{{.RepoPath}}/releases/download/v{{.Version}}/{{.Name}}-darwin-amd64.tar.gz"
      sha256 "{{.DarwinAMD64SHA}}"
    elsif Hardware::CPU.arm?
      url "https://github.com/{{.RepoPath}}/releases/download/v{{.Version}}/{{.Name}}-darwin-arm64.tar.gz"
      sha256 "{{.DarwinARM64SHA}}"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/{{.RepoPath}}/releases/download/v{{.Version}}/{{.Name}}-linux-amd64.tar.gz"
      sha256 "{{.LinuxAMD64SHA}}"
    elsif Hardware::CPU.arm?
      url "https://github.com/{{.RepoPath}}/releases/download/v{{.Version}}/{{.Name}}-linux-arm64.tar.gz"
      sha256 "{{.LinuxARM64SHA}}"
    end
  end

  def install
    binary = OS.mac? ? "{{.Name}}-darwin-" : "{{.Name}}-linux-"
    binary += Hardware::CPU.intel? ? "amd64" : "arm64"
    bin.install binary => "{{.Name}}"
  end

  test do
    assert_match "v{{.Version}}", shell_output("#{bin}/{{.Name}} --version")
  end
end
`
	data := struct {
		ClassName      string
		Description    string
		Homepage       string
		Version        string
		RepoPath       string
		Name           string
		DarwinAMD64SHA string
		DarwinARM64SHA string
		LinuxAMD64SHA  string
		LinuxARM64SHA  string
	}{
		ClassName:      className,
		Description:    p.Config.Description,
		Homepage:       p.Config.Homepage,
		Version:        version,
		RepoPath:       repoPath,
		Name:           p.Config.Name,
		DarwinAMD64SHA: resolved["darwin-amd64"],
		DarwinARM64SHA: resolved["darwin-arm64"],
		LinuxAMD64SHA:  resolved["linux-amd64"],
		LinuxARM64SHA:  resolved["linux-arm64"],
	}

	t, err := template.New("formula").Parse(formulaTmpl)
	if err != nil {
		return fmt.Errorf("parse formula template: %w", err)
	}

	f, err := os.Create(p.FormulaPath)
	if err != nil {
		return fmt.Errorf("create formula file: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("render formula: %w", err)
	}

	fmt.Printf("✅ Updated Homebrew formula: %s\n", p.FormulaPath)
	return nil
}

// UpdateAUR rewrites the AUR PKGBUILD with a new version and SHAs.
func UpdateAUR(p *AURPipeline, version string, sha256s map[string]string) error {
	if version == "" {
		version = getGitTagVersion(p.Config.Repository)
		if version == "" {
			version = p.Config.Version
		}
	}
	version = strings.TrimPrefix(version, "v")

	if _, err := os.Stat(p.AURDir); os.IsNotExist(err) {
		return fmt.Errorf("AUR repository not found at %s; run 'init aur' first", p.AURDir)
	}

	cleanRepo := strings.TrimSuffix(p.Config.Repository, "/")
	repoPath := strings.TrimPrefix(cleanRepo, "https://github.com/")

	platforms := []string{"linux-amd64", "linux-arm64"}

	resolved := make(map[string]string)
	for _, plat := range platforms {
		if sha256s != nil {
			if v, ok := sha256s[plat]; ok && v != "" {
				resolved[plat] = v
				continue
			}
		}
		url := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s-%s.tar.gz",
			repoPath, version, p.Config.Name, plat)
		fmt.Fprintf(os.Stderr, "📥 Downloading to calculate SHA256 for %s...\n", plat)
		sha, err := downloadAndHash(url)
		if err != nil {
			return fmt.Errorf("failed to resolve SHA256 for %s: %w", plat, err)
		}
		resolved[plat] = sha
	}

	pkgbuildTmpl := `# Maintainer: {{.Author}}
pkgname={{.PackageName}}
_binname={{.Name}}
pkgver={{.Version}}
pkgrel=1
pkgdesc="{{.Description}}"
arch=('x86_64' 'aarch64')
url="{{.Homepage}}"
license=('MIT')
depends=()

source_x86_64=("${_binname}-linux-amd64-${pkgver}.tar.gz::https://github.com/{{.RepoPath}}/releases/download/v${pkgver}/${_binname}-linux-amd64.tar.gz")
source_aarch64=("${_binname}-linux-arm64-${pkgver}.tar.gz::https://github.com/{{.RepoPath}}/releases/download/v${pkgver}/${_binname}-linux-arm64.tar.gz")
sha256sums_x86_64=('{{.LinuxAMD64SHA}}')
sha256sums_aarch64=('{{.LinuxARM64SHA}}')

package() {
  if [ "${CARCH}" = "x86_64" ]; then
    install -Dm755 "${srcdir}/${_binname}-linux-amd64" "${pkgdir}/usr/bin/${_binname}"
  elif [ "${CARCH}" = "aarch64" ]; then
    install -Dm755 "${srcdir}/${_binname}-linux-arm64" "${pkgdir}/usr/bin/${_binname}"
  fi
}
`
	data := struct {
		Author        string
		PackageName   string
		Name          string
		Version       string
		Description   string
		Homepage      string
		RepoPath      string
		LinuxAMD64SHA string
		LinuxARM64SHA string
	}{
		Author:        p.Config.Author,
		PackageName:   p.Config.GetAURPackageName(),
		Name:          p.Config.Name,
		Version:       version,
		Description:   p.Config.Description,
		Homepage:      p.Config.Homepage,
		RepoPath:      repoPath,
		LinuxAMD64SHA: resolved["linux-amd64"],
		LinuxARM64SHA: resolved["linux-arm64"],
	}

	t, err := template.New("pkgbuild").Parse(pkgbuildTmpl)
	if err != nil {
		return fmt.Errorf("parse PKGBUILD template: %w", err)
	}

	f, err := os.Create(p.PKGBUILDPath)
	if err != nil {
		return fmt.Errorf("create PKGBUILD: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("render PKGBUILD: %w", err)
	}

	fmt.Printf("✅ Updated AUR PKGBUILD: %s\n", p.PKGBUILDPath)
	return nil
}
