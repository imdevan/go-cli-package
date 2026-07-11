---
title: release
description: Build, tag, publish a GitHub release, and update package manifests
---

Build, tag, publish a GitHub release, and update package manifests

## Usage

```bash
go-cli-package release
```

## Flags




| Flag | Type | Description |
|------|------|-------------|
| --version | string | Release version (defaults to latest git tag, then package.toml version) |
| --sha256 | stringarray | Pre-computed SHA256s as platform=hash pairs, e.g. --sha256 linux-amd64=abc123 (repeatable; downloads if omitted) |
| --skip-tag | bool | Skip git tag creation |
| --skip-github | bool | Skip GitHub release creation (binaries won't be built) |
| --skip-update | bool | Skip Homebrew formula and AUR PKGBUILD updates |


## Source

See [release.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/release.go) for implementation details.
