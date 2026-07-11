---
title: release
description: Build, tag, and publish a GitHub release
---

Build, tag, and publish a GitHub release

## Usage

```bash
go-cli-package release [version]
```

## Flags




| Flag | Type | Description |
|------|------|-------------|
| --skip-tag | bool | Skip git tag creation |
| --skip-github | bool | Skip GitHub release creation (binaries won't be built) |


## Source

See [release.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/release.go) for implementation details.
