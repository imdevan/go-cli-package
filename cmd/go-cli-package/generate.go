package main

import (
	"go-cli-package/internal/workflow"

	"github.com/spf13/cobra"
)

// @docs-command:
//
//	name: generate
//	description:
//		Invokes the full docs generation pipeline:
//		1. Reads package metadata (TOML)
//		2. Generates markdown content pages
//		3. Parses Cobra commands
//		4. Generates command documentation
//		5. Generates API documentation (gomarkdoc)
//		6. Generates config (config.mjs)
//		7. Generates sidebar (sidebar.mjs)
//	example:
//		```bash
//		go-cli-docs generate
//		go-cli-docs generate --gen-api-docs=false
//		```
func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate all documentation from source",
		Long: `generate invokes the full docs generation pipeline:

  1. Read package metadata (TOML)
  2. Generate markdown content pages
  3. Parse Cobra commands
  4. Generate command documentation
  5. Generate API documentation (gomarkdoc)
  6. Generate config (config.mjs)
  7. Generate sidebar (sidebar.mjs)`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(genAPIDocs)
		},
	}

	return cmd
}

func runGenerate(genAPIDocs bool) error {
	return workflow.Generate(genAPIDocs, templatesOverride)
}
