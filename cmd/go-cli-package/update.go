package main

import (
	"fmt"
	"strings"

	"github.com/imdevan/go-cli-package/internal/pipeline"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var sha256Flags []string

	cmd := &cobra.Command{
		Use:   "update [homebrew|aur|all] [version]",
		Short: "Update package manifests with a new version and checksums",
		Long:  "Update Homebrew formula and/or AUR PKGBUILD with a new version and SHA256 checksums.",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := pipeline.FindAndLoadConfig(".")
			if err != nil {
				return err
			}
			pipelines := pipeline.NewPipelines(".", cfg)

			sha256s := parseSHA256Flags(sha256Flags)

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
				return pipeline.UpdateHomebrew(pipelines.Homebrew, resolved, sha256s)
			case "aur":
				return pipeline.UpdateAUR(pipelines.AUR, resolved, sha256s)
			case "all", "":
				if err := pipeline.UpdateHomebrew(pipelines.Homebrew, resolved, sha256s); err != nil {
					return fmt.Errorf("homebrew update failed: %w", err)
				}
				return pipeline.UpdateAUR(pipelines.AUR, resolved, sha256s)
			default:
				return fmt.Errorf("invalid target %q; must be 'homebrew', 'aur', or 'all'", target)
			}
		},
	}

	cmd.Flags().StringArrayVar(&sha256Flags, "sha256", nil,
		"SHA256 checksums as platform=hash pairs, e.g. --sha256 linux-amd64=abc123 (repeatable; downloads if omitted)")

	return cmd
}

// parseSHA256Flags converts []string{"linux-amd64=abc123", ...} into a map.
func parseSHA256Flags(flags []string) map[string]string {
	m := make(map[string]string, len(flags))
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}
