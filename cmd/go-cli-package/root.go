package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var rootCmd = newRootCmd()

// @docs-command:root
//
//	name: go-cli-package
//	description:
//		Package and release helper CLI.
//	example:
//		```bash
//		go-cli-package completion bash
//		```
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "go-cli-package",
		Short:        "Package and release helper CLI",
		SilenceUsage: true,
	}

	var showVersion bool

	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version and exit")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Println(resolvedVersion())
			return nil
		}
		return cmd.Help()
	}

	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newBuildCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeployCmd())
	cmd.AddCommand(newTagCmd())
	cmd.AddCommand(newReleaseCmd())
	cmd.AddCommand(newSyncCmd())

	return cmd
}

// resolvedVersion returns the build version embedded by go build, or "dev".
func resolvedVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		if v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

func Execute() error {
	return rootCmd.Execute()
}
