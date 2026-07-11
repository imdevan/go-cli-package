package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// pkgManagerCmds maps a package manager name to its create and install invocations.
// create: the full command string to scaffold Astro Starlight into a target directory.
// install: the command to install dependencies inside that directory.
type pmCmds struct {
	create  string // printf-formatted; receives docsDir
	install string // printf-formatted; receives docsDir
}

var pkgManagerCmds = map[string]pmCmds{
	"bun":  {"bun create astro@latest %s -- --template starlight --yes --no-install", "cd %s && bun install"},
	"npm":  {"npm create astro@latest %s -- --template starlight --yes --no-install", "cd %s && npm install --silent"},
	"yarn": {"yarn create astro@latest %s -- --template starlight --yes --no-install", "cd %s && yarn install --silent"},
	"pnpm": {"pnpm create astro@latest %s -- --template starlight --yes --no-install", "cd %s && pnpm install --silent"},
}

// @docs-command:
//
//	name: init
//	description:
//		Creates the docs/ folder and scaffolds Astro Starlight.
//		If the docs/ directory already exists, init is a no-op.
//		After scaffolding, init automatically runs generate to populate the docs site.
//	example:
//		```bash
//		go-cli-docs init
//		go-cli-docs init --pkg-manager npm
//		```
func newInitCmd() *cobra.Command {
	var pkgManager string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold the Astro Starlight docs directory",
		Long: `init creates the docs/ folder and scaffolds Astro Starlight (version-locked).

If docs/ already exists, init is a no-op.
After scaffolding, init automatically runs generate to populate the docs site.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(pkgManager, genAPIDocs)
		},
	}

	cmd.Flags().StringVarP(&pkgManager, "pkg-manager", "p", "bun", "Package manager to use (bun, npm, yarn, pnpm)")

	return cmd
}

func runInit(pkgManager string, genAPIDocs bool) error {
	docsDir := "docs"

	if _, err := os.Stat(docsDir); err == nil {
		fmt.Printf("✅ %s/ already exists – skipping init\n", docsDir)
		return nil
	}

	cmds, ok := pkgManagerCmds[pkgManager]
	if !ok {
		return fmt.Errorf("unsupported package manager %q (choose: bun, npm, yarn, pnpm)", pkgManager)
	}

	fmt.Printf("🔧 Initialising Astro Starlight in ./%s/ using %s...\n", docsDir, pkgManager)

	createScript := fmt.Sprintf(cmds.create, docsDir)
	initCmd := exec.Command("sh", "-c", createScript)
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("failed to init Astro Starlight: %w", err)
	}

	installScript := fmt.Sprintf(cmds.install, docsDir)
	installCmd := exec.Command("sh", "-c", installScript)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install Astro dependencies: %w", err)
	}

	fmt.Println("📝 Running initial generate...")
	return runGenerate(genAPIDocs)
}
