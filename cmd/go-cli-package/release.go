package main

import (
	"github.com/imdevan/go-cli-package/internal/pipeline"

	"github.com/spf13/cobra"
)

func newReleaseCmd() *cobra.Command {
	var (
		skipTag    bool
		skipGithub bool
	)

	cmd := &cobra.Command{
		Use:   "release [version]",
		Short: "Build, tag, and publish a GitHub release",
		Long: `End-to-end release automation that:
  1. Creates and pushes an annotated git tag.
  2. Builds cross-platform binaries and creates a GitHub release via 'gh'.

Use --skip-tag or --skip-github to omit individual steps.`,
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

			resolved, err := pipeline.ResolveVersion(version)
			if err != nil {
				return err
			}

			return pipeline.Release(".", pipelines, resolved, skipTag, skipGithub)
		},
	}

	cmd.Flags().BoolVar(&skipTag, "skip-tag", false, "Skip git tag creation")
	cmd.Flags().BoolVar(&skipGithub, "skip-github", false, "Skip GitHub release creation (binaries won't be built)")

	return cmd
}
