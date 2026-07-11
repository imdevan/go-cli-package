set shell := ["zsh", "-cu"]

CLI := "./bin/go-cli-package"

# Build

build:
	go build -o bin/go-cli-package ./cmd/go-cli-package
	@size=$(stat -c %s bin/go-cli-package 2>/dev/null || stat -f %z bin/go-cli-package 2>/dev/null); \
	echo "Build size: $(awk "BEGIN {printf \"%.2f MB\", $size/1048576}")"

build-run:
	go build -o bin/go-cli-package ./cmd/go-cli-package && ./bin/go-cli-package

watch:
	@rg --files | entr -r sh -c 'sleep 0.5; go build -o bin/go-cli-package ./cmd/go-cli-package'

dev-build:
	go build -gcflags "all=-N -l" -o bin/go-cli-package ./cmd/go-cli-package

install:
	install -m 0755 bin/go-cli-package /usr/local/bin/go-cli-package

uninstall:
	rm -f /usr/local/bin/go-cli-package

test:
	go test ./...

test-verbose:
	go test -v ./...

clean:
	rm -rf bin

# Documentation

docs-init args="":
	go-cli-docs init {{args}}

docs-generate args="":
	go-cli-docs generate {{args}}

docs-dev args="":
	go-cli-docs watch {{args}} & cd docs && bun install && bun run dev

docs-build: docs-generate
	@echo "🏗️  Building documentation site..."
	cd docs && NODE_ENV=production bun run build

docs-preview:
	@echo "👀 Previewing built documentation..."
	cd docs && bun run preview

docs-clean:
	@echo "🧹 Cleaning documentation build artifacts..."
	rm -rf docs/dist docs/.astro docs/node_modules docs/src/content/docs/api

# Pipeline init

init-homebrew-tap: build
	{{CLI}} init homebrew

init-aur-repo: build
	{{CLI}} init aur

init-all: build
	{{CLI}} init all

# Package updates

update-homebrew VERSION="": build
	{{CLI}} update homebrew {{VERSION}}

update-aur VERSION="": build
	{{CLI}} update aur {{VERSION}}

update VERSION="": build
	{{CLI}} update all {{VERSION}}

# Git tag management

tag-list: build
	{{CLI}} tag list

tag VERSION: build
	{{CLI}} tag create {{VERSION}}

tag-delete VERSION: build
	{{CLI}} tag delete {{VERSION}}

# Deploy

deploy-homebrew VERSION="": build
	{{CLI}} deploy homebrew {{VERSION}}

deploy-aur VERSION="": build
	{{CLI}} deploy aur {{VERSION}}

deploy-all VERSION="": build
	{{CLI}} deploy all {{VERSION}}

# aliases
publish-homebrew VERSION="": (deploy-homebrew VERSION)
publish-aur VERSION="": (deploy-aur VERSION)

# Release

release VERSION="": build
	{{CLI}} release {{VERSION}}

publish VERSION="": (release VERSION)
