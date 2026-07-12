package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// Helper to run a command in a directory
func runCmdInDir(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Check if a directory exists and handle reinitialization.
// Returns true if we should proceed with initialization, or false if the directory exists
// and force is false (indicating we should gracefully skip).
func checkAndPrepareDir(dir string, force bool) (bool, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return true, nil
	}

	if force {
		if err := os.RemoveAll(dir); err != nil {
			return false, fmt.Errorf("failed to remove existing directory %s: %w", dir, err)
		}
		return true, nil
	}

	return false, nil
}

// Extract GitHub user from repository URL
func extractGithubUser(repoURL string) string {
	cleanURL := strings.TrimSuffix(repoURL, "/")
	cleanURL = strings.TrimPrefix(cleanURL, "https://github.com/")
	cleanURL = strings.TrimPrefix(cleanURL, "git@github.com:")
	parts := strings.Split(cleanURL, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "username"
}

// InitializeHomebrew sets up the Homebrew tap directory, templates, and Git repo
func InitializeHomebrew(p *HomebrewPipeline, force bool) error {
	shouldProceed, err := checkAndPrepareDir(p.TapDir, force)
	if err != nil {
		return err
	}
	if !shouldProceed {
		fmt.Printf("⚠️  Homebrew tap already exists at %s; skipping initialization. Use --force to overwrite.\n", p.TapDir)
		return nil
	}

	githubUser := extractGithubUser(p.Config.Repository)

	// Try cloning first
	fmt.Printf("🍺 Attempting to clone Homebrew tap from GitHub...\n")
	remoteURL := fmt.Sprintf("git@github.com:%s/%s.git", githubUser, p.TapName)
	if err := runCmdInDir(".", "git", "clone", remoteURL, p.TapDir); err == nil {
		if _, err := os.Stat(filepath.Join(p.TapDir, "Formula")); err == nil {
			fmt.Printf("✅ Homebrew tap successfully cloned from GitHub to: %s\n", p.TapDir)
			return nil
		}
		fmt.Printf("ℹ️  Cloned repository is empty. Initializing files locally...\n")
	} else {
		// Clean up on failure to clone, to allow local creation
		_ = os.RemoveAll(p.TapDir)
		fmt.Printf("ℹ️  Clone failed (repository may not exist on GitHub yet). Initializing a new repository locally...\n")
	}

	fmt.Printf("🍺 Initializing Homebrew tap repository...\n")
	fmt.Printf("   Tap name: %s\n", p.TapName)
	fmt.Printf("   Package name: %s\n", p.Config.GetHomebrewPackageName())
	fmt.Printf("   Binary name: %s\n", p.Config.Name)
	fmt.Printf("   Location: %s\n", p.TapDir)

	formulaDir := filepath.Join(p.TapDir, "Formula")
	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		return fmt.Errorf("failed to create Formula directory: %w", err)
	}

	// Write README.md (with tildes replaced by backticks to prevent compilation issues)
	readmeTmpl := strings.ReplaceAll(`# Homebrew Tap for {{.PackageName}}

This is the official Homebrew tap for [{{.Name}}]({{.Homepage}}).

The formula is named ~{{.PackageName}}~ but installs the binary as ~{{.Name}}~.

## Installation

~~~bash
brew tap {{.GithubUser}}/{{.PackageName}}
brew install {{.PackageName}}
~~~

## Usage

After installation, use the ~{{.Name}}~ command:

~~~bash
{{.Name}} --help
~~~

## Updating

~~~bash
brew update
brew upgrade {{.PackageName}}
~~~

## Uninstall

~~~bash
brew uninstall {{.PackageName}}
brew untap {{.GithubUser}}/{{.PackageName}}
~~~
`, "~", "`")


	readmeData := struct {
		PackageName string
		Name        string
		Homepage    string
		GithubUser  string
	}{
		PackageName: p.Config.GetHomebrewPackageName(),
		Name:        p.Config.Name,
		Homepage:    p.Config.Homepage,
		GithubUser:  githubUser,
	}

	tReadme, err := template.New("readme").Parse(readmeTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse readme template: %w", err)
	}

	readmeFile, err := os.Create(filepath.Join(p.TapDir, "README.md"))
	if err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	defer readmeFile.Close()

	if err := tReadme.Execute(readmeFile, readmeData); err != nil {
		return fmt.Errorf("failed to execute readme template: %w", err)
	}

	// Write Formula template
	formulaTmpl := `class {{.ClassName}} < Formula
  desc "{{.Description}}"
  homepage "{{.Homepage}}"
  url "{{.RepoURL}}/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w", output: bin/"{{.Name}}"), "./cmd/{{.Name}}"
  end

  test do
    assert_match "v0.1.0", shell_output("#{bin}/{{.Name}} --version")
  end
end
`
	cleanRepo := strings.TrimSuffix(p.Config.Repository, "/")
	formulaData := struct {
		ClassName   string
		Description string
		Homepage    string
		RepoURL     string
		Name        string
	}{
		ClassName:   p.ClassName,
		Description: p.Config.Description,
		Homepage:    p.Config.Homepage,
		RepoURL:     cleanRepo,
		Name:        p.Config.Name,
	}

	tFormula, err := template.New("formula").Parse(formulaTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse formula template: %w", err)
	}

	formulaFile, err := os.Create(p.FormulaPath)
	if err != nil {
		return fmt.Errorf("failed to create Formula file: %w", err)
	}
	defer formulaFile.Close()

	if err := tFormula.Execute(formulaFile, formulaData); err != nil {
		return fmt.Errorf("failed to execute formula template: %w", err)
	}

	// Write .gitignore
	gitignoreContent := `.DS_Store
*.swp
*.swo
*~
`
	if err := os.WriteFile(filepath.Join(p.TapDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	gitDirExists := false
	if _, err := os.Stat(filepath.Join(p.TapDir, ".git")); err == nil {
		gitDirExists = true
	}

	// Git Init & Commit
	if !gitDirExists {
		if err := runCmdInDir(p.TapDir, "git", "init"); err != nil {
			return fmt.Errorf("failed to run git init: %w", err)
		}
		if err := runCmdInDir(p.TapDir, "git", "branch", "-M", "main"); err != nil {
			return fmt.Errorf("failed to rename branch to main: %w", err)
		}
	}
	if err := runCmdInDir(p.TapDir, "git", "add", "."); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}
	if err := runCmdInDir(p.TapDir, "git", "commit", "-m", fmt.Sprintf("Initial commit: Homebrew tap for %s", p.Config.GetHomebrewPackageName())); err != nil {
		return fmt.Errorf("failed to run git commit: %w", err)
	}

	repoCreated := false
	if gitDirExists {
		fmt.Printf("📤 Pushing initial commit to origin main...\n")
		if err := runCmdInDir(p.TapDir, "git", "push", "-u", "origin", "main"); err == nil {
			repoCreated = true
		} else {
			fmt.Printf("⚠️  Warning: failed to push to origin main: %v\n", err)
		}
	} else {
		if ghPath, err := exec.LookPath("gh"); err == nil {
			fmt.Printf("🚀 Creating GitHub repository %s/%s...\n", githubUser, p.TapName)
			cmd := exec.Command(ghPath, "repo", "create", fmt.Sprintf("%s/%s", githubUser, p.TapName), "--public")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("⚠️  Warning: failed to create GitHub repository via gh CLI: %v\n", err)
			} else {
				repoCreated = true
				_ = runCmdInDir(p.TapDir, "git", "remote", "add", "origin", remoteURL)
				fmt.Printf("📤 Pushing initial commit to origin main...\n")
				if err := runCmdInDir(p.TapDir, "git", "push", "-u", "origin", "main"); err != nil {
					fmt.Printf("⚠️  Warning: failed to push to origin main: %v\n", err)
				}
			}
		}
	}

	fmt.Printf("\n✅ Homebrew tap initialized at: %s\n\n", p.TapDir)
	if !repoCreated {
		fmt.Printf("Next steps:\n")
		fmt.Printf("1. Create a GitHub repository: https://github.com/new\n")
		fmt.Printf("   Repository name: %s\n", p.TapName)
		fmt.Printf("2. Push the tap:\n")
		fmt.Printf("   cd %s\n", p.TapDir)
		fmt.Printf("   git remote add origin git@github.com:%s/%s.git\n", githubUser, p.TapName)
		fmt.Printf("   git push -u origin main\n")
	} else {
		fmt.Printf("Next steps:\n")
		fmt.Printf("1. Update the formula with actual release SHA256 using:\n")
		fmt.Printf("   go-cli-package update [version]\n\n")
	}

	return nil
}

// InitializeAUR sets up the AUR repository directory, PKGBUILD, templates, and Git repo
func InitializeAUR(p *AURPipeline, force bool) error {
	shouldProceed, err := checkAndPrepareDir(p.AURDir, force)
	if err != nil {
		return err
	}
	if !shouldProceed {
		fmt.Printf("⚠️  AUR repository already exists at %s; skipping initialization. Use --force to overwrite.\n", p.AURDir)
		return nil
	}

	// Try cloning first
	fmt.Printf("📦 Attempting to clone AUR repository...\n")
	remoteURL := fmt.Sprintf("ssh://aur@aur.archlinux.org/%s.git", p.Config.GetAURPackageName())
	if err := runCmdInDir(".", "git", "clone", remoteURL, p.AURDir); err == nil {
		if _, err := os.Stat(p.PKGBUILDPath); err == nil {
			fmt.Printf("✅ AUR repository successfully cloned to: %s\n", p.AURDir)
			return nil
		}
		fmt.Printf("ℹ️  Cloned repository is empty. Initializing files locally...\n")
	} else {
		_ = os.RemoveAll(p.AURDir)
		fmt.Printf("ℹ️  Clone failed (repository may not exist on AUR yet). Initializing a new repository locally...\n")
	}

	fmt.Printf("📦 Initializing AUR repository...\n")
	fmt.Printf("   Package name: %s\n", p.Config.GetAURPackageName())
	fmt.Printf("   Binary name: %s\n", p.Config.Name)
	fmt.Printf("   Location: %s\n", p.AURDir)

	if err := os.MkdirAll(p.AURDir, 0755); err != nil {
		return fmt.Errorf("failed to create AUR directory: %w", err)
	}

	// Write PKGBUILD
	pkgbuildTmpl := `# Maintainer: {{.Author}}
pkgname={{.PackageName}}
_binname={{.Name}}
pkgver=0.1.0
pkgrel=1
pkgdesc="{{.Description}}"
arch=('x86_64' 'aarch64')
url="{{.Homepage}}"
license=('MIT')
depends=()
makedepends=('go')
source=("${_binname}-${pkgver}.tar.gz::{{.RepoURL}}/archive/refs/tags/v${pkgver}.tar.gz")
sha256sums=('REPLACE_WITH_ACTUAL_SHA256')

build() {
  cd "${_binname}-${pkgver}"
  export CGO_ENABLED=0
  export GOFLAGS="-buildmode=pie -trimpath -mod=readonly -modcacherw"
  go build -ldflags="-s -w" -o ${_binname} ./cmd/${_binname}
}

package() {
  cd "${_binname}-${pkgver}"
  install -Dm755 ${_binname} "${pkgdir}/usr/bin/${_binname}"
  if [ -f LICENSE ]; then
    install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
  fi
}
`
	cleanRepo := strings.TrimSuffix(p.Config.Repository, "/")
	pkgbuildData := struct {
		Author      string
		PackageName string
		Name        string
		Description string
		Homepage    string
		RepoURL     string
	}{
		Author:      p.Config.Author,
		PackageName: p.Config.GetAURPackageName(),
		Name:        p.Config.Name,
		Description: p.Config.Description,
		Homepage:    p.Config.Homepage,
		RepoURL:     cleanRepo,
	}

	tPkgbuild, err := template.New("pkgbuild").Parse(pkgbuildTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse PKGBUILD template: %w", err)
	}

	pkgbuildFile, err := os.Create(p.PKGBUILDPath)
	if err != nil {
		return fmt.Errorf("failed to create PKGBUILD file: %w", err)
	}
	defer pkgbuildFile.Close()

	if err := tPkgbuild.Execute(pkgbuildFile, pkgbuildData); err != nil {
		return fmt.Errorf("failed to execute PKGBUILD template: %w", err)
	}

	// Write README.md (with tildes replaced by backticks to prevent compilation issues)
	readmeTmpl := strings.ReplaceAll(`# AUR Package for {{.PackageName}}

This is the AUR (Arch User Repository) package for [{{.Name}}]({{.Homepage}}).

The package is named ~{{.PackageName}}~ but installs the binary as ~{{.Name}}~.

## Installation

### Using an AUR helper (recommended)

~~~bash
yay -S {{.PackageName}}
# or
paru -S {{.PackageName}}
~~~

### Manual installation

~~~bash
git clone https://aur.archlinux.org/{{.PackageName}}.git
cd {{.PackageName}}
makepkg -si
~~~

## Usage

After installation, use the ~{{.Name}}~ command:

~~~bash
{{.Name}} --help
~~~

## Updating

~~~bash
yay -Syu {{.PackageName}}
# or
paru -Syu {{.PackageName}}
~~~

## Uninstall

~~~bash
sudo pacman -R {{.PackageName}}
~~~

## Maintainer

{{.Author}}
`, "~", "`")

	readmeData := struct {
		PackageName string
		Name        string
		Homepage    string
		Author      string
	}{
		PackageName: p.Config.GetAURPackageName(),
		Name:        p.Config.Name,
		Homepage:    p.Config.Homepage,
		Author:      p.Config.Author,
	}

	tReadme, err := template.New("readme").Parse(readmeTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse README template: %w", err)
	}

	readmeFile, err := os.Create(filepath.Join(p.AURDir, "README.md"))
	if err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}
	defer readmeFile.Close()

	if err := tReadme.Execute(readmeFile, readmeData); err != nil {
		return fmt.Errorf("failed to execute README template: %w", err)
	}

	// Write .gitignore
	gitignoreContent := `*.tar.gz
*.tar.xz
*.zip
pkg/
src/
*.pkg.tar.*
`
	if err := os.WriteFile(filepath.Join(p.AURDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	// Generate .SRCINFO (requires makepkg)
	makepkgPath, err := exec.LookPath("makepkg")
	if err != nil {
		fmt.Printf("⚠️  Warning: makepkg not found in PATH; required to generate .SRCINFO\n")
	} else {
		srcinfoPath := filepath.Join(p.AURDir, ".SRCINFO")
		srcinfo, err := os.Create(srcinfoPath)
		if err != nil {
			return fmt.Errorf("create .SRCINFO: %w", err)
		}
		cmd := exec.Command(makepkgPath, "--printsrcinfo")
		cmd.Dir = p.AURDir
		cmd.Stdout = srcinfo
		cmd.Stderr = os.Stderr
		runErr := cmd.Run()
		srcinfo.Close()
		if runErr != nil {
			return fmt.Errorf("makepkg --printsrcinfo: %w", runErr)
		}
	}

	gitDirExists := false
	if _, err := os.Stat(filepath.Join(p.AURDir, ".git")); err == nil {
		gitDirExists = true
	}

	// Git Init & Commit
	if !gitDirExists {
		if err := runCmdInDir(p.AURDir, "git", "init"); err != nil {
			return fmt.Errorf("failed to run git init: %w", err)
		}
		if err := runCmdInDir(p.AURDir, "git", "branch", "-M", "master"); err != nil {
			return fmt.Errorf("failed to rename branch to master: %w", err)
		}
	}
	
	gitAddArgs := []string{"add", "PKGBUILD", "README.md", ".gitignore"}
	if _, err := os.Stat(filepath.Join(p.AURDir, ".SRCINFO")); err == nil {
		gitAddArgs = append(gitAddArgs, ".SRCINFO")
	}
	if err := runCmdInDir(p.AURDir, "git", gitAddArgs...); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}
	if err := runCmdInDir(p.AURDir, "git", "commit", "-m", fmt.Sprintf("Initial commit: AUR package for %s", p.Config.GetAURPackageName())); err != nil {
		return fmt.Errorf("failed to run git commit: %w", err)
	}

	repoCreated := false

	if !gitDirExists {
		fmt.Printf("🚀 Registering AUR remote: %s\n", remoteURL)
		_ = runCmdInDir(p.AURDir, "git", "remote", "add", "origin", remoteURL)
	}

	fmt.Printf("📤 Pushing initial commit to AUR master...\n")
	if err := runCmdInDir(p.AURDir, "git", "push", "-u", "origin", "master"); err != nil {
		fmt.Printf("⚠️  Warning: failed to push to AUR remote (does the package exist or is your SSH key correct?): %v\n", err)
	} else {
		repoCreated = true
	}

	fmt.Printf("\n✅ AUR repository initialized at: %s\n\n", p.AURDir)
	if !repoCreated {
		fmt.Printf("Next steps:\n")
		fmt.Printf("1. Register an AUR account: https://aur.archlinux.org/register\n")
		fmt.Printf("2. Add your SSH key to AUR: https://aur.archlinux.org/account\n")
		fmt.Printf("3. Push the package manually:\n")
		fmt.Printf("   cd %s\n", p.AURDir)
		fmt.Printf("   git push -u origin master\n")
	} else {
		fmt.Printf("Next steps:\n")
		fmt.Printf("1. Update PKGBUILD with actual release SHA256 using:\n")
		fmt.Printf("   go-cli-package update [version]\n\n")
	}

	return nil
}
