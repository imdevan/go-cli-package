package workflow

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"go-cli-package/internal/templates"
	"github.com/pelletier/go-toml/v2"
)

// PackageInfo holds values from internal/package/package.toml
type PackageInfo struct {
	Name        string `toml:"name"`
	PackageName string `toml:"package_name"`
	Module      string `toml:"module"`
	Description string `toml:"description"`
	Short       string `toml:"short"`
	Version     string `toml:"version"`
	Homepage    string `toml:"homepage"`
	Repository  string `toml:"repository"`
	Author      string `toml:"author"`
	DocsSite    string `toml:"docs_site"`
	DocsBase    string `toml:"docs_base"`
}

// GenerateDocs performs the full documentation generation using Go templates.
// templatesOverride is a list of files and/or directories checked for
// user-supplied templates before falling back to the embedded defaults.
func GenerateDocs(genAPIDocs bool, templatesOverride []string) error {
	// 1. Load package metadata (TOML)
	pkgInfo, err := loadPackageInfo()
	if err != nil {
		return fmt.Errorf("failed to load package info: %w", err)
	}

	// 2. Parse Cobra commands
	commands, err := parseCommands("./cmd/" + pkgInfo.Name)
	if err != nil {
		return fmt.Errorf("failed to parse commands: %w", err)
	}

	for i := range commands {
		if commands[i].CmdName == "" {
			words := strings.Fields(commands[i].Use)
			if len(words) > 0 {
				commands[i].CmdName = words[0]
			} else {
				base := filepath.Base(commands[i].GoFile)
				commands[i].CmdName = strings.TrimSuffix(base, filepath.Ext(base))
			}
		}
		if commands[i].CmdName == "root" || filepath.Base(commands[i].GoFile) == "root.go" || strings.Contains(strings.ToLower(commands[i].VarName), "rootcmd") {
			commands[i].CmdName = pkgInfo.Name
		}
		commands[i].Path = []string{commands[i].CmdName}
	}

	relations, relErr := parseCommandRelations("./cmd/" + pkgInfo.Name)
	if relErr == nil {
		varNameToCmdName := make(map[string]string)
		for _, c := range commands {
			varNameToCmdName[c.VarName] = c.CmdName
		}

		for i := range commands {
			if commands[i].CmdName == pkgInfo.Name {
				continue
			}

			var chain []string
			current := commands[i].VarName
			for {
				parent, exists := relations[current]
				if !exists {
					break
				}
				chain = append([]string{parent}, chain...)
				current = parent
			}

			if len(chain) > 0 {
				var nameParts []string
				for _, parentVar := range chain {
					if parentCmdName, ok := varNameToCmdName[parentVar]; ok {
						if parentCmdName != pkgInfo.Name {
							nameParts = append(nameParts, parentCmdName)
						}
					}
				}
				nameParts = append(nameParts, commands[i].CmdName)
				commands[i].Path = nameParts
				commands[i].CmdName = strings.Join(nameParts, "-")
			}
		}
	}

	// 3. Generate config (docs/config.mjs)
	if err := writeConfigMjs(pkgInfo, templatesOverride); err != nil {
		return err
	}

	// 4. Generate sidebar files (docs/sidebar.mjs)
	if err := writeSidebarMjs(pkgInfo, commands, templatesOverride); err != nil {
		return err
	}

	// 5. Update docs/astro.config.mjs to use dynamic header and config
	if err := writeAstroConfigMjs(templatesOverride); err != nil {
		return err
	}

	// 6. Generate base custom.css if it does not already exist
	if err := writeCustomCss(pkgInfo, templatesOverride); err != nil {
		return err
	}

	// 7. Strip landing page if it exists (Astro Starlight default index.mdx)
	indexMdx := filepath.Join("docs", "src", "content", "docs", "index.mdx")
	if _, err := os.Stat(indexMdx); err == nil {
		_ = os.Remove(indexMdx)
	}

	// 7. Generate markdown content pages
	if err := generateContentPage("README.md", "index.md", pkgInfo.Name, pkgInfo.Description, templatesOverride); err != nil {
		return err
	}
	if err := generateContentPage("INSTALL.md", "install.md", "Install", fmt.Sprintf("Installation instructions for %s", pkgInfo.Name), templatesOverride); err != nil {
		return err
	}
	if err := generateContentPage("CONFIG.md", "configuration.md", "Configuration", fmt.Sprintf("Configuration options for %s", pkgInfo.Name), templatesOverride); err != nil {
		return err
	}
	// CONTRIBUTING.md is optional
	_ = generateContentPage("CONTRIBUTING.md", "contributing.md", "Contributing", fmt.Sprintf("Contributing to %s", pkgInfo.Name), templatesOverride)

	// 8. Generate command documentation
	if err := generateCommandPages(pkgInfo, commands, templatesOverride); err != nil {
		return err
	}

	// 9. Generate API documentation via gomarkdoc
	apiDir := filepath.Join("docs", "src", "content", "docs", "api")
	if genAPIDocs {
		if err := generateAPIDocs(pkgInfo); err != nil {
			return err
		}
	} else {
		// Remove existing api folder to keep directory clean
		if _, err := os.Stat(apiDir); err == nil {
			_ = os.RemoveAll(apiDir)
		}
	}

	fmt.Println("✅ Documentation generation completed successfully!")
	return nil
}

func loadPackageInfo() (*PackageInfo, error) {
	pkgFile := filepath.Join("internal", "package", "package.toml")
	data, err := ioutil.ReadFile(pkgFile)
	if err != nil {
		return nil, err
	}
	var info PackageInfo
	if err := toml.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	if info.Name == "" {
		cwd, _ := os.Getwd()
		info.Name = filepath.Base(cwd)
	}
	return &info, nil
}

func writeConfigMjs(info *PackageInfo, templatesOverride []string) error {
	tmpl, err := templates.Load("config.mjs.tmpl", templatesOverride)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	data := struct {
		DocsSite    string
		DocsBase    string
		Repository  string
		ProjectName string
		Description string
	}{
		DocsSite:    info.DocsSite,
		DocsBase:    info.DocsBase,
		Repository:  info.Repository,
		ProjectName: info.Name,
		Description: info.Description,
	}

	path := filepath.Join("docs", "config.mjs")
	if err := templates.RenderToFile(tmpl, path, data); err != nil {
		return fmt.Errorf("failed to write config.mjs: %w", err)
	}
	return nil
}

func writeSidebarMjs(info *PackageInfo, commands []CommandInfo, templatesOverride []string) error {
	tmpl, err := templates.Load("sidebar.mjs.tmpl", templatesOverride)
	if err != nil {
		return fmt.Errorf("failed to parse sidebar template: %w", err)
	}

	tree := buildSidebarTree(commands, info.Name)
	commandsJS := formatSidebarItems(tree, "    ")

	apiPackages, err := detectAPIPackages()
	if err != nil {
		return err
	}

	data := struct {
		ProjectName string
		CommandsJS  string
		APIPackages []string
	}{
		ProjectName: info.Name,
		CommandsJS:  commandsJS,
		APIPackages: apiPackages,
	}

	path := filepath.Join("docs", "sidebar.mjs")
	if err := templates.RenderToFile(tmpl, path, data); err != nil {
		return fmt.Errorf("failed to write sidebar.mjs: %w", err)
	}
	return nil
}

func writeAstroConfigMjs(templatesOverride []string) error {
	tmpl, err := templates.Load("astro.config.mjs.tmpl", templatesOverride)
	if err != nil {
		return fmt.Errorf("failed to parse astro config template: %w", err)
	}

	path := filepath.Join("docs", "astro.config.mjs")
	if err := templates.RenderToFile(tmpl, path, nil); err != nil {
		return fmt.Errorf("failed to write astro.config.mjs: %w", err)
	}
	return nil
}

func writeCustomCss(info *PackageInfo, templatesOverride []string) error {
	path := filepath.Join("docs", "src", "styles", "custom.css")
	if _, err := os.Stat(path); err == nil {
		// File already exists, do not overwrite
		return nil
	}

	// Create directory if it does not exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpl, err := templates.Load("custom.css.tmpl", templatesOverride)
	if err != nil {
		return fmt.Errorf("failed to parse custom css template: %w", err)
	}

	data := struct {
		ProjectName string
	}{
		ProjectName: info.Name,
	}

	if err := templates.RenderToFile(tmpl, path, data); err != nil {
		return fmt.Errorf("failed to write custom.css: %w", err)
	}
	return nil
}

func detectAPIPackages() ([]string, error) {
	internalPath := "internal"
	entries, err := os.ReadDir(internalPath)
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		// skip special dirs
		if name == "testutil" || name == "adapters" || name == "package" || name == "templates" {
			continue
		}
		// check for .go files
		files, _ := os.ReadDir(filepath.Join(internalPath, name))
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
				pkgs = append(pkgs, name)
				break
			}
		}
	}
	return pkgs, nil
}

// extractMarkdownTitle checks if the first non-empty line of content is an H1 header.
// If it is, it returns the title extracted from the header and the content with that header line removed.
// Otherwise, it returns the defaultTitle and the original content.
func extractMarkdownTitle(content string, defaultTitle string) (string, string) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		if strings.HasPrefix(trimmedLine, "#") {
			rest := trimmedLine[1:]
			if len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t') {
				titleVal := strings.TrimSpace(rest)
				titleVal = strings.TrimRight(titleVal, "#")
				title := strings.TrimSpace(titleVal)

				var filtered []string
				filtered = append(filtered, lines[:i]...)
				filtered = append(filtered, lines[i+1:]...)
				return title, strings.Join(filtered, "\n")
			}
		}
		break
	}
	return defaultTitle, content
}

func generateContentPage(srcFile, dstFile, title, description string, templatesOverride []string) error {
	if _, err := os.Stat(srcFile); err != nil {
		return nil // optional source missing
	}
	data, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}

	title, content := extractMarkdownTitle(string(data), title)

	tmpl, err := templates.Load("page.md.tmpl", templatesOverride)
	if err != nil {
		return fmt.Errorf("failed to parse page template: %w", err)
	}

	outDir := filepath.Join("docs", "src", "content", "docs")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	dstPath := filepath.Join(outDir, dstFile)

	templateData := struct {
		Title       string
		Description string
		Content     string
	}{
		Title:       title,
		Description: description,
		Content:     content,
	}

	if err := templates.RenderToFile(tmpl, dstPath, templateData); err != nil {
		return fmt.Errorf("failed to write content page %s: %w", dstFile, err)
	}
	return nil
}

type TemplateFlag struct {
	Display     string
	Type        string
	Description string
}

type TemplateFlagGroup struct {
	Name        string
	Description string
	Example     string
	Note        string
	Flags       []TemplateFlag
}

type TemplateSubcommand struct {
	Name        string
	Description string
}

type CommandTemplateData struct {
	Name        string
	Description string
	Doc         string
	Usage       string
	FlagGroups  []TemplateFlagGroup
	Subcommands []TemplateSubcommand
	Repository  string
	SourceFile  string
	SourcePath  string
}

func generateCommandPages(pkgInfo *PackageInfo, commands []CommandInfo, templatesOverride []string) error {
	outDir := filepath.Join("docs", "src", "content", "docs", "commands")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	for _, c := range commands {
		if err := writeCommandPage(outDir, &c, pkgInfo, commands, templatesOverride); err != nil {
			return err
		}
	}
	return nil
}

func writeCommandPage(dir string, cmdInfo *CommandInfo, pkgInfo *PackageInfo, allCmds []CommandInfo, templatesOverride []string) error {
	tmpl, err := templates.Load("command.md.tmpl", templatesOverride)
	if err != nil {
		return fmt.Errorf("failed to parse command template: %w", err)
	}

	// Prepare Flag Groups
	var flagGroups []TemplateFlagGroup
	for _, fg := range cmdInfo.FlagGroups {
		var flags []TemplateFlag
		for _, f := range fg.Flags {
			display := "--" + f.Name
			if f.Short != "" && f.Short != "null" {
				display = fmt.Sprintf("-%s, --%s", f.Short, f.Name)
			}
			flags = append(flags, TemplateFlag{
				Display:     display,
				Type:        strings.ToLower(f.Type),
				Description: f.Description,
			})
		}
		flagGroups = append(flagGroups, TemplateFlagGroup{
			Name:        fg.Name,
			Description: fg.Description,
			Example:     fg.Example,
			Note:        fg.Note,
			Flags:       flags,
		})
	}

	// Inherit persistent flags from the root command into subcommand docs,
	// mirroring cobra's own "Global Flags" section in --help output.
	if cmdInfo.CmdName != pkgInfo.Name {
		ownFlags := make(map[string]bool)
		for _, fg := range cmdInfo.FlagGroups {
			for _, f := range fg.Flags {
				ownFlags[f.Name] = true
			}
		}

		var globalFlags []TemplateFlag
		for _, c := range allCmds {
			if c.CmdName != pkgInfo.Name {
				continue
			}
			for _, fg := range c.FlagGroups {
				for _, f := range fg.Flags {
					if !f.Persistent || ownFlags[f.Name] {
						continue
					}
					display := "--" + f.Name
					if f.Short != "" && f.Short != "null" {
						display = fmt.Sprintf("-%s, --%s", f.Short, f.Name)
					}
					globalFlags = append(globalFlags, TemplateFlag{
						Display:     display,
						Type:        strings.ToLower(f.Type),
						Description: f.Description,
					})
				}
			}
		}

		if len(globalFlags) > 0 {
			flagGroups = append(flagGroups, TemplateFlagGroup{
				Name:  "Global Flags",
				Flags: globalFlags,
			})
		}
	}

	// Prepare Subcommands (only for root command)
	var subcommands []TemplateSubcommand
	if cmdInfo.CmdName == pkgInfo.Name {
		for _, c := range allCmds {
			if c.CmdName == pkgInfo.Name {
				continue
			}
			desc := c.Short
			if desc == "" || desc == "null" {
				desc = c.CmdName
			}
			subcommands = append(subcommands, TemplateSubcommand{
				Name:        c.CmdName,
				Description: desc,
			})
		}
	}

	// Prepare Usage
	var usage string
	if cmdInfo.CmdName == pkgInfo.Name {
		usage = fmt.Sprintf("```bash\n%s\n```", pkgInfo.Name)
	} else {
		parentPath := ""
		if len(cmdInfo.Path) > 1 {
			parentPath = strings.Join(cmdInfo.Path[:len(cmdInfo.Path)-1], " ")
		}

		useVal := cmdInfo.Use
		if useVal == "" || useVal == "null" {
			useVal = strings.Join(cmdInfo.Path, " ")
		} else if parentPath != "" {
			if !strings.HasPrefix(useVal, parentPath) {
				useVal = parentPath + " " + useVal
			}
		}
		usage = fmt.Sprintf("```bash\n%s %s\n```", pkgInfo.Name, useVal)
	}

	// Prepare Description
	desc := cmdInfo.Short
	if desc == "" || desc == "null" {
		if cmdInfo.CmdName == pkgInfo.Name {
			desc = pkgInfo.Description
		} else {
			desc = cmdInfo.Short
		}
	}
	if desc == "" || desc == "null" {
		desc = pkgInfo.Description
	}

	docText := cmdInfo.Doc
	if docText == "" || docText == "null" {
		if cmdInfo.CmdName == pkgInfo.Name {
			docText = pkgInfo.Description
		} else {
			docText = cmdInfo.Short
		}
	}

	// Source file details
	sourceFile := filepath.Base(cmdInfo.GoFile)
	sourcePath := filepath.Join("cmd", pkgInfo.Name, sourceFile)

	data := CommandTemplateData{
		Name:        strings.Join(cmdInfo.Path, " "),
		Description: desc,
		Doc:         docText,
		Usage:       usage,
		FlagGroups:  flagGroups,
		Subcommands: subcommands,
		Repository:  strings.TrimSuffix(pkgInfo.Repository, "/"),
		SourceFile:  sourceFile,
		SourcePath:  sourcePath,
	}

	dstPath := filepath.Join(dir, fmt.Sprintf("%s.md", cmdInfo.CmdName))
	if err := templates.RenderToFile(tmpl, dstPath, data); err != nil {
		return fmt.Errorf("failed to write command page %s: %w", cmdInfo.CmdName, err)
	}
	return nil
}

func stripGomarkdocHeader(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			if i+1 < len(lines) {
				return strings.Join(lines[i+1:], "\n")
			}
			return ""
		}
	}
	return content
}

func generateAPIDocs(pkgInfo *PackageInfo) error {
	// Ensure gomarkdoc is available
	if _, err := exec.LookPath("gomarkdoc"); err != nil {
		fmt.Println("Installing gomarkdoc...")
		install := exec.Command("go", "install", "github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest")
		install.Stdout = os.Stdout
		install.Stderr = os.Stderr
		if err := install.Run(); err != nil {
			return fmt.Errorf("failed to install gomarkdoc: %w", err)
		}
	}

	apiDir := filepath.Join("docs", "src", "content", "docs", "api")
	tmpDir, err := ioutil.TempDir("", "api_tmp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	pkgs, err := detectAPIPackages()
	if err != nil {
		return err
	}

	cleanRepository := strings.TrimSuffix(pkgInfo.Repository, "/")

	for _, pkg := range pkgs {
		pkgPath := "./" + filepath.Join("internal", pkg)
		outFile := filepath.Join(tmpDir, pkg+".raw.md")
		footer := fmt.Sprintf("## Source\n\nSee [internal/%s/](%s/blob/main/internal/%s/) for implementation details.", pkg, cleanRepository, pkg)

		cmd := exec.Command("gomarkdoc",
			"--output", outFile,
			"--footer", footer,
			pkgPath,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  No exported symbols in %s\n", pkg)
			continue
		}

		rawBytes, _ := os.ReadFile(outFile)
		strippedContent := stripGomarkdocHeader(string(rawBytes))

		final := fmt.Sprintf("---\ntitle: %s\ndescription: API documentation for the %s package\n---\n\n%s", pkg, pkg, strippedContent)
		dstPath := filepath.Join(apiDir, pkg+".md")
		os.MkdirAll(apiDir, 0755)
		os.WriteFile(dstPath, []byte(final), 0644)
	}

	// adapters
	adaptersPath := filepath.Join("internal", "adapters")
	if entries, _ := os.ReadDir(adaptersPath); len(entries) > 0 {
		adaptersOut := filepath.Join(apiDir, "adapters")
		os.MkdirAll(adaptersOut, 0755)
		for _, a := range entries {
			if !a.IsDir() {
				continue
			}
			name := a.Name()
			pkgPath := "./" + filepath.Join(adaptersPath, name)
			outFile := filepath.Join(tmpDir, name+".raw.md")
			footer := fmt.Sprintf("## Source\n\nSee [internal/adapters/%s/](%s/blob/main/internal/adapters/%s/) for implementation details.", name, cleanRepository, name)

			cmd := exec.Command("gomarkdoc",
				"--output", outFile,
				"--footer", footer,
				pkgPath,
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("⚠️  No exported symbols in adapter %s\n", name)
				continue
			}

			rawBytes, _ := os.ReadFile(outFile)
			strippedContent := stripGomarkdocHeader(string(rawBytes))

			final := fmt.Sprintf("---\ntitle: adapters/%s\ndescription: API documentation for the %s adapter\n---\n\n%s", name, name, strippedContent)
			dstPath := filepath.Join(adaptersOut, name+".md")
			os.WriteFile(dstPath, []byte(final), 0644)
		}
	}
	return nil
}

type SidebarItem struct {
	Label string
	Link  string
	Items []*SidebarItem
}

func buildSidebarTree(commands []CommandInfo, rootName string) []*SidebarItem {
	var rootItems []*SidebarItem

	findItem := func(items []*SidebarItem, label string) *SidebarItem {
		for _, item := range items {
			if item.Label == label {
				return item
			}
		}
		return nil
	}

	sortedCmds := make([]CommandInfo, len(commands))
	copy(sortedCmds, commands)
	sort.Slice(sortedCmds, func(i, j int) bool {
		return sortedCmds[i].CmdName < sortedCmds[j].CmdName
	})

	for _, c := range sortedCmds {
		if c.CmdName == rootName {
			continue // Skip root command
		}

		parts := c.Path
		
		currentLevel := &rootItems
		var currentPath []string

		for i, part := range parts {
			currentPath = append(currentPath, part)
			label := strings.Join(currentPath, " ")
			isLast := (i == len(parts)-1)

			if isLast {
				existing := findItem(*currentLevel, part)
				if existing != nil {
					existing.Link = "/commands/" + c.CmdName
				} else {
					*currentLevel = append(*currentLevel, &SidebarItem{
						Label: label,
						Link:  "/commands/" + c.CmdName,
					})
				}
			} else {
				existing := findItem(*currentLevel, part)
				if existing == nil {
					newGroup := &SidebarItem{
						Label: part,
					}
					*currentLevel = append(*currentLevel, newGroup)
					existing = newGroup
				} else if existing.Link != "" && len(existing.Items) == 0 {
					selfCopy := &SidebarItem{
						Label: existing.Label,
						Link:  existing.Link,
					}
					existing.Link = ""
					existing.Items = append(existing.Items, selfCopy)
				}
				currentLevel = &existing.Items
			}
		}
	}
	return rootItems
}

func formatSidebarItems(items []*SidebarItem, indent string) string {
	var sb strings.Builder
	for _, item := range items {
		if len(item.Items) > 0 {
			sb.WriteString(fmt.Sprintf("%s{\n", indent))
			sb.WriteString(fmt.Sprintf("%s  label: '%s',\n", indent, item.Label))
			sb.WriteString(fmt.Sprintf("%s  items: [\n", indent))
			sb.WriteString(formatSidebarItems(item.Items, indent+"    "))
			sb.WriteString(fmt.Sprintf("%s  ],\n", indent))
			sb.WriteString(fmt.Sprintf("%s},\n", indent))
		} else {
			sb.WriteString(fmt.Sprintf("%s{ label: '%s', link: '%s' },\n", indent, item.Label, item.Link))
		}
	}
	return sb.String()
}
