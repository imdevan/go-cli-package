---
title: Install
description: Installation instructions for go-cli-package
---

## Recommended: 

### Start with [go-cli-template](http://devan.gg/go-cli-template/#quick-start)

I would strongly recommended checking out my go-cli-template with everything including docs setup!

I made this project to distribute the docs config for that project and it's children.
### Just add this to your justfile

```bash
_install-go-cli-docs:
    GOBIN=$(PWD)/bin go install "github.com/imdevan/go-cli-docs/cmd/go-cli-docs@v1.0.0"
      
# Documentation tasks
docs-init args="": _install-go-cli-docs
	./bin/go-cli-docs init {{args}}

docs-generate args="": _install-go-cli-docs
	./bin/go-cli-docs generate {{args}}

docs-dev args="": _install-go-cli-docs
	./bin/go-cli-docs watch {{args}} & cd docs && bun install && bun run dev
```

This will install the package in bin for use with your project. 

```bash
just docs-init      # init docs
just docs-generate  # generate docs
just docs-dev       # watch for changes and run dev server
```

## Global use

### homebrew
```bash
brew install imdevan/go-cli-docs/go-cli-docs
```

### arch (aur)
```bash
yay -s go-cli-docs
```

### github release

download the latest binary for your platform from the [releases page](https://github.com/imdevan/go-cli-docs/releases).

```bash
# linux (amd64)
curl -l https://github.com/imdevan/go-cli-docs/releases/latest/download/go-cli-template-linux-amd64.tar.gz | tar -xz
sudo mv go-cli-docs-linux-amd64 /usr/local/bin/go-cli-template
```

```bash
# macos (apple silicon)
curl -l https://github.com/imdevan/go-cli-docs/releases/latest/download/go-cli-template-darwin-arm64.tar.gz | tar -xz
sudo mv go-cli-docs-darwin-arm64 /usr/local/bin/go-cli-template
```

```bash
# macos (intel)
curl -l https://github.com/imdevan/go-cli-docs/releases/latest/download/go-cli-template-darwin-amd64.tar.gz | tar -xz
sudo mv go-cli-docs-darwin-amd64 /usr/local/bin/go-cli-template
```

## manual
```bash
gh repo clone imdevan/go-cli-docs
cd go-cli-docs
just build
sudo just install
```

