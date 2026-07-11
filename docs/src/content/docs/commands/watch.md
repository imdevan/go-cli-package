---
title: watch
description: Watch source files and re-generate documentation on change
---

Monitors source files for changes and automatically re-runs generate.
Watched patterns: *.md, *.go, package.toml
Excluded paths: node_modules/, docs/src/content/docs/, .git/

### Example

```bash
go-cli-docs watch
go-cli-docs watch --gen-api-docs=false
```

## Usage

```bash
go-cli-package watch
```

## Source

See [watch.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/watch.go) for implementation details.
