package main

import (
	"github.com/imdevan/go-cli-package/internal/pipeline"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var configFlag string

	cmd := &cobra.Command{
		Use:   "sync [config-path]",
		Short: "Sync project files with package.toml configuration",
		Long:  "Synchronize module path, imports, binary name, cmd directory name, completion examples, and description across files based on package.toml.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := ""
			if len(args) > 0 {
				configPath = args[0]
			} else if configFlag != "" {
				configPath = configFlag
			}

			return pipeline.Sync(".", configPath)
		},
	}

	cmd.Flags().StringVarP(&configFlag, "config", "c", "", "Path to package.toml configuration file")

	return cmd
}
