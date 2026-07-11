package main

import (
	"fmt"

	"go-cli-package/internal/pipeline"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag [list|create|delete]",
		Short: "Manage git tags",
		Long:  "List, create, or delete annotated git tags and keep the remote in sync.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sub := "list"
			if len(args) > 0 {
				sub = args[0]
			}
			switch sub {
			case "list":
				return cmd.Help()
			default:
				return fmt.Errorf("unknown subcommand %q; use 'list', 'create', or 'delete'", sub)
			}
		},
	}

	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagCreateCmd())
	cmd.AddCommand(newTagDeleteCmd())

	return cmd
}

func newTagListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available git tags",
		Long:  "Print all git tags sorted by version (newest first).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pipeline.TagList(".", limit)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of tags to display (0 = all)")

	return cmd
}

func newTagCreateCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "create [version]",
		Short: "Create and push an annotated git tag",
		Long:  "Create an annotated git tag vVERSION and push it to origin. Use --force to overwrite an existing tag.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) > 0 {
				version = args[0]
			}
			resolved, err := pipeline.ResolveVersion(version)
			if err != nil {
				return err
			}
			return pipeline.TagCreate(".", resolved, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Delete and recreate the tag if it already exists")

	return cmd
}

func newTagDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [version]",
		Short: "Delete a git tag locally and remotely",
		Long:  "Delete tag vVERSION both locally and from the origin remote.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) > 0 {
				version = args[0]
			}
			resolved, err := pipeline.ResolveVersion(version)
			if err != nil {
				return err
			}
			return pipeline.TagDelete(".", resolved)
		},
	}

	return cmd
}
