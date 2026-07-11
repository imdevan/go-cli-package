---
title: build
description: Build package targets (Go binary and/or AUR PKGBUILD)
---

Build package targets (Go binary and/or AUR PKGBUILD)

## Usage

```bash
go-cli-package build [binary|aur|all] [version]
```

## Flags




| Flag | Type | Description |
|------|------|-------------|
| --sha256 | string | SHA256 checksum of the source archive (required for AUR) |


## Source

See [build.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/build.go) for implementation details.
