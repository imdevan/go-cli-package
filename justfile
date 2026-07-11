set shell := ["zsh", "-cu"]

PACKAGE := "go-cli-package"
PACKAGE_BIN := "./bin/" + PACKAGE
PACKAGE_CMD := "./cmd/" + PACKAGE

# Build

build:
	go build -o {{PACKAGE_BIN}} {{PACKAGE_CMD}}
	@size=$(stat -c %s {{PACKAGE_BIN}} 2>/dev/null || stat -f %z {{PACKAGE_BIN}} 2>/dev/null); \
	echo "Build size: $(awk "BEGIN {printf \"%.2f MB\", $size/1048576}")"

build-run:
	go build -o {{PACKAGE_BIN}} {{PACKAGE_CMD}} && {{PACKAGE_BIN}}

watch:
	@rg --files | entr -r sh -c 'sleep 0.5; go build -o {{PACKAGE_BIN}} {{PACKAGE_CMD}}'

dev-build:
	go build -gcflags "all=-N -l" -o {{PACKAGE_BIN}} {{PACKAGE_CMD}}

install:
	install -m 0755 {{PACKAGE_BIN}} /usr/local/bin/{{PACKAGE}}

uninstall:
	rm -f /usr/local/bin/{{PACKAGE}}

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

# Github release management
# ================================================================================

# Github tag management

tag-list: build
	{{PACKAGE_BIN}} tag list

tag version: build
	{{PACKAGE_BIN}} tag create {{version}}

tag-delete version: build
	{{PACKAGE_BIN}} tag delete {{version}}

# Github release

release version="": build
	{{PACKAGE_BIN}} release {{version}}

publish version="": (release version)

# Package management
# ================================================================================

# Pipeline init

init-homebrew-tap: build
	{{PACKAGE_BIN}} init homebrew

init-aur-repo: build
	{{PACKAGE_BIN}} init aur

init-all: build
	{{PACKAGE_BIN}} init all

# Package updates

update-homebrew version="": build
	{{PACKAGE_BIN}} update homebrew {{version}}

update-aur version="": build
	{{PACKAGE_BIN}} update aur {{version}}

update version="": build
	{{PACKAGE_BIN}} update all {{version}}


# Deploy

deploy-homebrew version="": build
	{{PACKAGE_BIN}} deploy homebrew {{version}}

deploy-aur version="": build
	{{PACKAGE_BIN}} deploy aur {{version}}

deploy-all version="": build
	{{PACKAGE_BIN}} deploy all {{version}}

# aliases
publish-homebrew version="": (deploy-homebrew version)
publish-aur version="": (deploy-aur version)


