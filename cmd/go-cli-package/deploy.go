package main

import (
	"fmt"

	"go-cli-package/internal/pipeline"
	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy [homebrew|aur|all] [version]",
		Short: "Push updated package manifests to their remotes",
		Long:  "Commit and push the updated Homebrew formula and/or AUR PKGBUILD to their respective remotes.",
		Args:  cobra.MaximumNArgs(2),
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
			version := ""
			if len(args) > 1 {
				version = args[1]
			}

			resolved, err := pipeline.ResolveVersion(version)
			if err != nil {
				return err
			}

			switch target {
			case "homebrew":
				return pipeline.DeployHomebrew(pipelines.Homebrew, resolved)
			case "aur":
				return pipeline.DeployAUR(pipelines.AUR, resolved)
			case "all", "":
				if err := pipeline.DeployHomebrew(pipelines.Homebrew, resolved); err != nil {
					return fmt.Errorf("homebrew deploy failed: %w", err)
				}
				return pipeline.DeployAUR(pipelines.AUR, resolved)
			default:
				return fmt.Errorf("invalid target %q; must be 'homebrew', 'aur', or 'all'", target)
			}
		},
	}

	return cmd
}
