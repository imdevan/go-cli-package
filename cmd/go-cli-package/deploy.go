package main

import (
	"fmt"

	"go-cli-package/internal/pipeline"
	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "deploy [homebrew|aur|all]",
		Short: "Push updated package manifests to their remotes",
		Long:  "Commit and push the updated Homebrew formula and/or AUR PKGBUILD to their respective remotes.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := pipeline.FindAndLoadConfig(".")
			if err != nil {
				return err
			}
			pipelines := pipeline.NewPipelines(".", cfg)

			target := "all"
			if len(args) > 0 {
				target = args[0]
			}

			switch target {
			case "homebrew":
				return pipeline.DeployHomebrew(pipelines.Homebrew, version)
			case "aur":
				return pipeline.DeployAUR(pipelines.AUR, version)
			case "all", "":
				if err := pipeline.DeployHomebrew(pipelines.Homebrew, version); err != nil {
					return fmt.Errorf("homebrew deploy failed: %w", err)
				}
				return pipeline.DeployAUR(pipelines.AUR, version)
			default:
				return fmt.Errorf("invalid target %q; must be 'homebrew', 'aur', or 'all'", target)
			}
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Release version (defaults to package.toml version)")

	return cmd
}
