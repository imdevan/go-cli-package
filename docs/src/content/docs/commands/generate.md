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

## Source

See [generate.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/generate.go) for implementation details.
