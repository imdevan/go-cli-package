---
title: go-cli-package
description: Generate Astro Starlight documentation for Go CLI projects
---

Generate Astro Starlight documentation for Go CLI projects.
The tool parses Cobra commands and flags, rendering markdown pages,
sidebar configs, and API docs.

### Example

```bash
go-cli-docs init
go-cli-docs generate
go-cli-docs watch
```

## Usage

```bash
go-cli-package
```

## Flags




| Flag | Type | Description |
|------|------|-------------|
| -v, --version | bool | Print version and exit |
| -a, --gen-api-docs | bool | Generate API documentation via gomarkdoc |
| -t, --templates | stringarray | Path to a file or directory of custom templates overriding the embedded defaults (repeatable) |


## Available Commands


- [`completion`](/commands/completion) - Generate shell completion scripts
- [`generate`](/commands/generate) - Generate all documentation from source
- [`init`](/commands/init) - Scaffold the Astro Starlight docs directory
- [`watch`](/commands/watch) - Watch source files and re-generate documentation on change

## Source

See [root.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/root.go) for implementation details.
