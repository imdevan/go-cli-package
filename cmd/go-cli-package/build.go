package main

import (
	"fmt"

	"go-cli-package/internal/pipeline"
	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var version string
	var sha256 string

	cmd := &cobra.Command{
		Use:   "build [binary|aur|all]",
		Short: "Build package targets (Go binary and/or AUR PKGBUILD)",
		Long:  "Build package targets. By default, both the local Go binary and the AUR PKGBUILD are built/generated.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := pipeline.FindAndLoadConfig(".")
			if err != nil {
				return err
			}

			target := "all"
			if len(args) > 0 {
				target = args[0]
			}

			switch target {
			case "binary":
				return pipeline.BuildBinary(".", cfg)
			case "aur":
				return pipeline.BuildAUR(".", cfg, version, sha256)
			case "all", "":
				if err := pipeline.BuildBinary(".", cfg); err != nil {
					return fmt.Errorf("binary build failed: %w", err)
				}
				if err := pipeline.BuildAUR(".", cfg, version, sha256); err != nil {
					return fmt.Errorf("aur build failed: %w", err)
				}
				return nil
			default:
				return fmt.Errorf("invalid build target %q; must be 'binary', 'aur', or 'all'", target)
			}
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Version of the release (defaults to git tag or package.toml version)")
	cmd.Flags().StringVar(&sha256, "sha256", "", "SHA256 checksum of the source archive (required for AUR)")

	return cmd
}
