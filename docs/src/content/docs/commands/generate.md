---
title: generate
description: Generate all documentation from source
---

Invokes the full docs generation pipeline:
1. Reads package metadata (TOML)
2. Generates markdown content pages
3. Parses Cobra commands
4. Generates command documentation
5. Generates API documentation (gomarkdoc)
6. Generates config (config.mjs)
7. Generates sidebar (sidebar.mjs)

### Example

```bash
go-cli-docs generate
go-cli-docs generate --gen-api-docs=false
```

## Usage

```bash
go-cli-package generate
```

## Flags

### Global Flags



| Flag | Type | Description |
|------|------|-------------|
| -a, --gen-api-docs | bool | Generate API documentation via gomarkdoc |
| -t, --templates | stringarray | Path to a file or directory of custom templates overriding the embedded defaults (repeatable) |


## Source

See [generate.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/generate.go) for implementation details.
