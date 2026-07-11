package main

import (
	"fmt"

	"go-cli-package/internal/pipeline"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init [homebrew|aur|all]",
		Short: "Initialize packaging repositories (Homebrew tap and/or AUR repository)",
		Long:  "Initialize packaging repositories for the project. By default, both Homebrew tap and AUR repository are initialized.",
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
				return pipeline.InitializeHomebrew(pipelines.Homebrew, force)
			case "aur":
				return pipeline.InitializeAUR(pipelines.AUR, force)
			case "all", "":
				if err := pipeline.InitializeHomebrew(pipelines.Homebrew, force); err != nil {
					return fmt.Errorf("homebrew tap init failed: %w", err)
				}
				if err := pipeline.InitializeAUR(pipelines.AUR, force); err != nil {
					return fmt.Errorf("aur repo init failed: %w", err)
				}
				return nil
			default:
				return fmt.Errorf("invalid init target %q; must be 'homebrew', 'aur', or 'all'", target)
			}
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force reinitialization, overwriting existing directories")

	return cmd
}
