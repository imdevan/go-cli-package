---
title: go-cli-package
description: Package and release helper CLI
---

Package and release helper CLI.

### Example

```bash
go-cli-package completion bash
```

## Usage

```bash
go-cli-package
```

## Flags




| Flag | Type | Description |
|------|------|-------------|
| -v, --version | bool | Print version and exit |


## Available Commands


- [`build`](/commands/build) - Build package targets (Go binary and/or AUR PKGBUILD)
- [`completion`](/commands/completion) - Generate shell completion scripts
- [`init`](/commands/init) - Initialize packaging repositories (Homebrew tap and/or AUR repository)

## Source

See [root.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/root.go) for implementation details.
