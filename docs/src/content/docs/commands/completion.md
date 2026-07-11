---
title: completion
description: Generate shell completion scripts
---

Generate shell completion scripts for bash, zsh, fish, or powershell.

### Example

```bash
go-cli-package completion bash > /etc/bash_completion.d/go-cli-package
```

## Usage

```bash
go-cli-package completion [bash|zsh|fish|powershell]
```

## Source

See [completion.go](https://github.com/imdevan/go-cli-package/blob/main/cmd/go-cli-package/completion.go) for implementation details.
