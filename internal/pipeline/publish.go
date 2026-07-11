package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PublishHomebrew commits and pushes the updated Homebrew formula to the tap remote.
func PublishHomebrew(p *HomebrewPipeline, version string) error {
	if version == "" {
		version = p.Config.Version
	}
	version = strings.TrimPrefix(version, "v")

	if _, err := os.Stat(p.TapDir); os.IsNotExist(err) {
		return fmt.Errorf("homebrew tap not found at %s; run 'init homebrew' first", p.TapDir)
	}

	fmt.Printf("🍺 Publishing Homebrew formula for v%s...\n", version)

	pkgName := p.Config.GetPackageName()
	githubUser := extractGithubUser(p.Config.Repository)

	// Ensure origin remote exists
	if err := ensureRemote(p.TapDir, "origin",
		fmt.Sprintf("git@github.com:%s/%s.git", githubUser, p.TapName)); err != nil {
		return err
	}

	// Stage formula file
	if err := runCmdInDir(p.TapDir, "git", "add", fmt.Sprintf("Formula/%s.rb", pkgName)); err != nil {
		return fmt.Errorf("git add formula: %w", err)
	}

	// Commit if there are staged changes
	if err := commitIfChanged(p.TapDir, fmt.Sprintf("Update %s to v%s", pkgName, version)); err != nil {
		return err
	}

	// Push, setting upstream on first push
	if err := pushWithUpstream(p.TapDir, "origin", "main"); err != nil {
		return fmt.Errorf("git push homebrew tap: %w", err)
	}

	fmt.Printf("✅ Homebrew formula published!\n")
	fmt.Printf("   Install with: brew tap %s/%s && brew install %s\n", githubUser, pkgName, pkgName)
	return nil
}

// PublishAUR commits and pushes the updated PKGBUILD to the AUR remote.
func PublishAUR(p *AURPipeline, version string) error {
	if version == "" {
		version = p.Config.Version
	}
	version = strings.TrimPrefix(version, "v")

	if _, err := os.Stat(p.AURDir); os.IsNotExist(err) {
		return fmt.Errorf("AUR repository not found at %s; run 'init aur' first", p.AURDir)
	}

	fmt.Printf("📦 Publishing AUR package for v%s...\n", version)

	pkgName := p.Config.GetPackageName()

	// Generate .SRCINFO (requires makepkg)
	makepkgPath, err := exec.LookPath("makepkg")
	if err != nil {
		return fmt.Errorf("makepkg not found in PATH; required to generate .SRCINFO")
	}

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

	// Ensure AUR remote
	if err := ensureRemote(p.AURDir, "origin",
		fmt.Sprintf("ssh://aur@aur.archlinux.org/%s.git", pkgName)); err != nil {
		return err
	}

	if err := runCmdInDir(p.AURDir, "git", "add", "PKGBUILD", ".SRCINFO"); err != nil {
		return fmt.Errorf("git add PKGBUILD .SRCINFO: %w", err)
	}

	if err := commitIfChanged(p.AURDir, fmt.Sprintf("Update %s to v%s", pkgName, version)); err != nil {
		return err
	}

	if err := pushWithUpstream(p.AURDir, "origin", "master"); err != nil {
		return fmt.Errorf("git push AUR: %w", err)
	}

	fmt.Printf("✅ AUR package published!\n")
	fmt.Printf("   https://aur.archlinux.org/packages/%s\n", pkgName)
	return nil
}

// ensureRemote adds a git remote if not already present.
func ensureRemote(dir, name, url string) error {
	cmd := exec.Command("git", "remote", "get-url", name)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		fmt.Printf("Adding remote %s → %s\n", name, url)
		return runCmdInDir(dir, "git", "remote", "add", name, url)
	}
	return nil
}

// commitIfChanged commits staged changes only if there are any.
func commitIfChanged(dir, message string) error {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// Non-zero exit means there are staged changes
		return runCmdInDir(dir, "git", "commit", "-m", message)
	}
	fmt.Println("Nothing to commit — already up to date.")
	return nil
}

// pushWithUpstream pushes to the remote, setting upstream on first push.
func pushWithUpstream(dir, remote, branch string) error {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// No upstream set — first push
		return runCmdInDir(dir, "git", "push", "-u", remote, branch)
	}
	return runCmdInDir(dir, "git", "push")
}
