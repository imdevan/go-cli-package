package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the metadata defined in package.toml
type Config struct {
	Name                string `toml:"name"`
	PackageName         string `toml:"package_name"`
	HomebrewPackageName string `toml:"homebrew_package_name"`
	AURPackageName      string `toml:"aur_package_name"`
	Module              string `toml:"module"`
	Description         string `toml:"description"`
	Short               string `toml:"short"`
	Version             string `toml:"version"`
	Homepage            string `toml:"homepage"`
	Repository          string `toml:"repository"`
	Author              string `toml:"author"`
	DocsSite            string `toml:"docs_site"`
	DocsBase            string `toml:"docs_base"`
}

// GetPackageName returns PackageName or Name if PackageName is empty.
func (c *Config) GetPackageName() string {
	if c.PackageName != "" {
		return c.PackageName
	}
	return c.Name
}

// GetHomebrewPackageName returns HomebrewPackageName if not empty, otherwise GetPackageName().
func (c *Config) GetHomebrewPackageName() string {
	if c.HomebrewPackageName != "" {
		return c.HomebrewPackageName
	}
	return c.GetPackageName()
}

// GetAURPackageName returns AURPackageName if not empty, otherwise GetPackageName().
func (c *Config) GetAURPackageName() string {
	if c.AURPackageName != "" {
		return c.AURPackageName
	}
	return c.GetPackageName()
}

// HomebrewPipeline represents the Homebrew delivery channel.
type HomebrewPipeline struct {
	Config      *Config
	TapName     string
	TapDir      string
	FormulaPath string
	ClassName   string
}

// AURPipeline represents the Arch User Repository delivery channel.
type AURPipeline struct {
	Config       *Config
	AURDir       string
	PKGBUILDPath string
}

// Pipelines holds the initialized target package pipelines.
type Pipelines struct {
	RootDir  string
	Config   *Config
	Homebrew *HomebrewPipeline
	AUR      *AURPipeline
}

// LoadConfig reads the TOML configuration from the given file path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML config: %w", err)
	}

	if cfg.Name == "" {
		return nil, fmt.Errorf("name field is required in config")
	}

	return &cfg, nil
}

// FindAndLoadConfig searches for package.toml in common locations under rootDir.
func FindAndLoadConfig(rootDir string) (*Config, error) {
	paths := []string{
		filepath.Join(rootDir, "internal", "package", "package.toml"),
		filepath.Join(rootDir, "package.toml"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return LoadConfig(p)
		}
	}

	return nil, fmt.Errorf("package.toml configuration file not found in common locations under %s", rootDir)
}

// NewPipelines creates a Pipelines representation from the root directory and config.
func NewPipelines(rootDir string, cfg *Config) *Pipelines {
	hbPkgName := cfg.GetHomebrewPackageName()
	aurPkgName := cfg.GetAURPackageName()
	
	// Homebrew fields
	tapName := "homebrew-" + hbPkgName
	tapDir := filepath.Join(rootDir, tapName)
	formulaPath := filepath.Join(tapDir, "Formula", hbPkgName+".rb")
	
	// CamelCase conversion for Homebrew formula class name
	className := toCamelCase(hbPkgName)
	
	// AUR fields
	aurDir := filepath.Join(rootDir, "aur-"+aurPkgName)
	aurPKGBUILDPath := filepath.Join(aurDir, "PKGBUILD")

	return &Pipelines{
		RootDir: rootDir,
		Config:  cfg,
		Homebrew: &HomebrewPipeline{
			Config:      cfg,
			TapName:     tapName,
			TapDir:      tapDir,
			FormulaPath: formulaPath,
			ClassName:   className,
		},
		AUR: &AURPipeline{
			Config:       cfg,
			AURDir:       aurDir,
			PKGBUILDPath: aurPKGBUILDPath,
		},
	}
}

func toCamelCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, "")
}
