package pipeline

import (
	"fmt"
)

// ResolveVersion returns the provided version if not empty, otherwise reads the version
// from package.toml in the current directory.
func ResolveVersion(provided string) (string, error) {
	if provided != "" {
		return provided, nil
	}

	cfg, err := FindAndLoadConfig(".")
	if err != nil {
		return "", fmt.Errorf("failed to resolve version: %w", err)
	}

	if cfg.Version == "" {
		return "", fmt.Errorf("version field is empty in package.toml")
	}

	return cfg.Version, nil
}
