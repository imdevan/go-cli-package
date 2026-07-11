package main

import (
	"go-cli-package/internal/pipeline"

	"github.com/spf13/cobra"
)

func newReleaseCmd() *cobra.Command {
	var (
		sha256Flags []string
		skipTag     bool
		skipGithub  bool
		skipUpdate  bool
	)

	cmd := &cobra.Command{
		Use:   "release [version]",
		Short: "Build, tag, publish a GitHub release, and update package manifests",
		Long: `End-to-end release automation that:
  1. Creates and pushes an annotated git tag.
  2. Builds cross-platform binaries and creates a GitHub release via 'gh'.
  3. Updates the Homebrew formula and AUR PKGBUILD.

Use --skip-tag, --skip-github, or --skip-update to omit individual steps.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := pipeline.FindAndLoadConfig(".")
			if err != nil {
				return err
			}
			pipelines := pipeline.NewPipelines(".", cfg)

			version := ""
			if len(args) > 0 {
				version = args[0]
			}

			sha256s := parseSHA256Flags(sha256Flags)

			return pipeline.Release(".", pipelines, version, sha256s, skipTag, skipGithub, skipUpdate)
		},
	}

	cmd.Flags().StringArrayVar(&sha256Flags, "sha256", nil,
		"Pre-computed SHA256s as platform=hash pairs, e.g. --sha256 linux-amd64=abc123 (repeatable; downloads if omitted)")
	cmd.Flags().BoolVar(&skipTag, "skip-tag", false, "Skip git tag creation")
	cmd.Flags().BoolVar(&skipGithub, "skip-github", false, "Skip GitHub release creation (binaries won't be built)")
	cmd.Flags().BoolVar(&skipUpdate, "skip-update", false, "Skip Homebrew formula and AUR PKGBUILD updates")

	return cmd
}
