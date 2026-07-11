---
title: update
description: Update package manifests with a new version and checksums
---

Update package manifests with a new version and checksums

## Usage

```bash
go-cli-package update [homebrew|aur|all] [version]
```

## Flags




| Flag | Type | Description |
|------|------|-------------|
| --sha256 | stringarray | SHA256 checksums as platform=hash pairs, e.g. --sha256 linux-amd64=abc123 (repeatable; downloads if omitted) |


## Source

See [update.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/update.go) for implementation details.
