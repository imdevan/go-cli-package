---
title: go-cli-package
description: A go cli to deploy your go cli package.
---


A go cli package to manage your go cli package deployment.

## Features

- manage GitHub tags and releases
- init, update, and deploy aur and homebrew packages
- easily callable from make or just build tools. 
- doesn't add to your go cli build size

## Author's Note

This project was created by deconstructing another cli tool I created: [prompter-cli](https://devan.gg/prompter-cli), a cli tool to organize prompts and skills. I wanted a reusable way to create go cli tools that were fast and looked pretty.

I then used this project to bootstrap [bookmark](https://devan.gg/bookmark), a go based approach to organize shell aliases. I then took what I learned building bookmark and incorporated those learnings back into this project. 

If you're here; those projects may also interest you! :)


## Requirements

- [Go]([https://go.dev/]) for doing the thing.
- [Just](https://just.systems/) for running scripts.
- [Bun](https://bun.sh/) for docs generation. Easily sub for `npm` if preferred.

## Quick start

```bash
# Clone the repo
gh repo clone imdevan/go-cli-docs
cd go-cli-docs

# Build the binary
just build

# Scaffold a new Astro Starlight docs site (inside your target project)
go-cli-docs init

# Generate docs from source
go-cli-docs generate

# Watch for changes and regenerate automatically
go-cli-docs watch
```

## Using this as a template

This project is designed to be forked and adapted. The single source of truth for all project metadata is `internal/package/package.toml`. 
<br>
Changing it and running `just sync` propagates your values everywhere.

### Steps

1. Fork or clone the repository.
2. Edit `internal/package/package.toml` with your project details:

```toml
name        = "my-tool"
module      = "github.com/you/my-tool"
description = "What my tool does"
version     = "0.1.0"
repository  = "https://github.com/you/my-tool"
docs_site   = "https://you.github.io"
docs_base   = "/my-tool"
```

3. Run `just sync` to propagate changes.
4. Review the diff with `git diff`.
5. Build and verify: `just build && just test`

### What `just sync` updates

- Go module name in `go.mod` and all import paths throughout `internal/` and `cmd/`
- Binary name in the justfile and build scripts
- Shell completion examples
- README description block
- Version constant in `cmd/*/root.go`

After syncing, add your own commands under `cmd/`, domain types under `internal/domain/`, and UI components under `internal/ui/`.

## Documentation

Documentation is built with [Astro Starlight](https://starlight.astro.build/) and lives in `docs/`. Content is generated automatically from the Go source and project markdown files — you generally don't edit `docs/src/content/docs/` by hand.

However you can easily customize the look of your docs by editing the styles located in `docs/src/styles/custom.css`.

### Go Docs

Go docs are generated via [gomarkdoc](https://github.com/princjef/gomarkdoc)

Into **API Reference**

These docs are intended to be be seen only for the development of the project.

They contain the documentation for the development of the project; they are not needed by users of the cli tool you are building.

### readme, install, config, and contributing

These pages are pulled whole sale from the markdown files in the project. 
Frontmatter is added to play nice with Startlight.

### Commands `cmd`

These pages are generated from go doc comments from the `/cmd` folder which
contains all commands a potential user of the cli tool would use. 

As well as some bash scripting to pull out information on any flag params for a given command.

### How docs are generated

#### just docs-dev

Build the docs and watch for changes.

#### just docs-build

Build docs for production.

#### just docs-generate

Running `just docs-generate` (or implicitly via `just docs-dev` / `just docs-build`) runs `scripts/docs_generate.sh`, which:

1. **Reads `internal/package/package.toml`** and writes `docs/config.mjs` and `docs/sidebar.mjs` with the current project name, description, repository URL, and base path.
2. **Imports markdown files** from the repository root:
   - `README.md` → `docs/src/content/docs/index.md`
   - `INSTALL.md` → `docs/src/content/docs/install.md`
   - `CONFIG.md` → `docs/src/content/docs/configuration.md`
3. **Generates command pages** by parsing each `cmd/<name>/*.go` file for `Use`, `Short`, flags, and godoc comments — one page per command under `docs/src/content/docs/commands/`.
4. **Generates API reference pages** using [gomarkdoc](https://github.com/princjef/gomarkdoc) for every package under `internal/` (including `internal/adapters/*`), outputting to `docs/src/content/docs/api/`.
    - The API reference is only rendered in the side bar during development. 
    - As it is internal and not pertinent to users of the cli tool, but still very helpful for maintainers

### API Reference visibility

The API Reference section is **internal** and only rendered in development by default. When you run `just docs-dev`, the sidebar includes all `internal/` package docs. In a production build (`NODE_ENV=production`), the API Reference is hidden — unless the project name is `go-cli-template`, in which case it is always shown as a live example.

This means when you use this template for your own project, your production docs site will be clean and user-facing, while you still get the full API reference locally during development.

```bash
just docs-dev      # Serves docs at http://localhost:4321 with API reference visible
just docs-build    # Builds production site — API reference hidden for non-template projects
```

## Architecture

Items marked `*` are updated by `just sync`.

```
.
├── go.mod                      # Go packages       *
├── justfile                    # Just run commands *
├── README.md                   # You are here      *
│
├── cmd/
│   └── go-cli-docs/
│       ├── main.go             # Binary entry point
│       ├── root.go             # Root command & global flags
│       ├── init.go             # `init` sub-command  — scaffold Astro Starlight
│       ├── generate.go         # `generate` sub-command — run docs pipeline
│       ├── watch.go            # `watch` sub-command  — fsnotify file watcher
│       └── completion.go       # Shell completion sub-command *
│
└── internal/
    ├── package/
    │   └── package.toml        # Source of truth — edit this, then run just sync
    │
    ├── app/                    # Application bootstrap
    ├── domain/                 # Core types and data models
    ├── errors/                 # Shared error types
    ├── templates/              # Embedded Go templates (*.tmpl)
    ├── workflow/               # Docs-generation pipeline logic
    ├── utils/                  # Stateless helpers
    │   ├── paths.go            # XDG config/data/cache path resolution
    │   └── time.go             # Time formatting helpers
    └── testutil/               # Shared test fixtures and helpers
```

## Commands

### `go-cli-docs init`

Scaffolds an Astro Starlight `docs/` directory inside your project, installs dependencies, and runs an initial `generate`.

```bash
# Scaffold with the default package manager (bun)
go-cli-docs init

# Use a different package manager
go-cli-docs init --pkg-manager npm
go-cli-docs init --pkg-manager yarn
go-cli-docs init --pkg-manager pnpm
```

> If `docs/` already exists, `init` is a no-op — safe to re-run.

### `go-cli-docs generate`

Runs the full documentation-generation pipeline:

1. Reads `internal/package/package.toml` metadata
2. Generates markdown content pages (README, INSTALL, CONFIG, CONTRIBUTING)
3. Parses Cobra command definitions and writes command pages
4. Generates API reference pages via [gomarkdoc](https://github.com/princjef/gomarkdoc)
5. Writes `docs/config.mjs` and `docs/sidebar.mjs`

```bash
# Generate all docs (including API reference)
go-cli-docs generate

# Skip gomarkdoc API docs generation
go-cli-docs generate --gen-api-docs=false
```

### `go-cli-docs watch`

Monitors source files for changes and automatically re-runs `generate`. Run the Astro dev server in a separate terminal for a live-preview workflow.

Watched patterns: `*.md`, `*.go`, `*.toml`  
Excluded paths: `node_modules/`, `docs/src/content/docs/`, `.git/`

```bash
# Watch and regenerate on every source change
go-cli-docs watch

# Watch without regenerating API reference docs
go-cli-docs watch --gen-api-docs=false

# Full live-preview workflow (two terminals)
go-cli-docs watch          # terminal 1 — watches & regenerates
cd docs && bun run dev     # terminal 2 — Astro dev server

# Or use the justfile shortcut (runs both in one command)
just docs-dev
```

### Shell completion

```bash
go-cli-docs completion bash   # bash
go-cli-docs completion zsh    # zsh
go-cli-docs completion fish   # fish
go-cli-docs completion powershell  # PowerShell
```

## Development

### Build & Run

```bash
just build           # Build the binary
just build-run       # Build and run the binary
just dev-build       # Build with debug symbols (disables optimizations)
just watch           # Watch for changes and rebuild automatically
just install         # Install binary to /usr/local/bin
just uninstall       # Remove binary from /usr/local/bin
just clean           # Remove build artifacts (bin/)
```

### Testing

```bash
just test            # Run all tests
just test-verbose    # Run tests with verbose output
```

### Project Sync

```bash
just sync            # Sync all project files from package.toml metadata
```

### Documentation

```bash
just docs-init       # Scaffold Astro Starlight docs/ (runs go-cli-docs init)
just docs-generate   # Generate all docs from source (runs go-cli-docs generate)
just docs-dev        # Watch for changes + start Astro dev server
just docs-build      # Generate docs and build production site
just docs-preview    # Preview the production build locally
just docs-clean      # Remove generated docs and build artifacts
```

### Package Distribution

```bash
just init-homebrew-tap            # Initialize a Homebrew tap repository
just init-aur-repo                # Initialize an AUR repository
just update-homebrew-formula 1.0.0  # Update Homebrew formula to a version
just update-aur-pkgbuild 1.0.0      # Update AUR PKGBUILD to a version
```

### Tags & Releases

```bash
just tag 1.0.0           # Create and push a git tag
just tag-delete 1.0.0    # Delete a git tag locally and remotely
just tag-list            # List recent tags
just release 1.0.0       # Full release (build, tag, publish)
just github-release 1.0.0   # Create a GitHub release with assets
just deploy-aur 1.0.0       # Deploy to AUR
just deploy-homebrew 1.0.0  # Deploy to Homebrew tap
just deploy-all 1.0.0       # Deploy to all targets
```

## Global Flags

All three commands (`init`, `generate`, `watch`) share these persistent root-level flags:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--gen-api-docs` | `-a` | bool | `true` | Generate API reference via gomarkdoc |
| `--templates` | `-t` | string (repeatable) | — | Path to a file or directory of custom templates overriding the embedded defaults |
| `--version` | `-v` | bool | `false` | Print version and exit |

```bash
# Disable API docs generation for faster runs
go-cli-docs generate --gen-api-docs=false
go-cli-docs watch --gen-api-docs=false

# Use one or more custom template overrides
go-cli-docs generate --templates ./my-templates
go-cli-docs generate -t ./override1 -t ./override2

# Print version
go-cli-docs --version
```

## Project Metadata

The single source of truth for project metadata is `internal/package/package.toml`.
Edit it and re-run `go-cli-docs generate` to propagate changes across all generated docs.

```toml
name        = "my-tool"
module      = "github.com/you/my-tool"
description = "What my tool does"
version     = "0.1.0"
repository  = "https://github.com/you/my-tool"
docs_site   = "https://you.github.io"
docs_base   = "/my-tool"
```

See `CONFIG.md` for the full reference.

## Installation

See `INSTALL.md` for installation options.

# Thank you!

This project was made by deconstructing another cli project of mine [Prompter](http://devan.gg/prompter-cli/). Check it out if you like fiddling with coding agents and want a more vim centric way of managing your prompting!

