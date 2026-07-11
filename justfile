set shell := ["zsh", "-cu"]

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

build-aur:
	./scripts/build_aur.sh

install:
	install -m 0755 bin/go-cli-package /usr/local/bin/go-cli-package

uninstall:
	rm -f /usr/local/bin/go-cli-package

test:
	go test ./...

test-verbose:
	go test -v ./...

sync:
	./scripts/sync.sh

clean:
	rm -rf bin

# Documentation tasks
docs-init args="":
	./bin/go-cli-package init {{args}}

docs-generate args="":
	./bin/go-cli-package generate {{args}}

docs-dev args="":
	./bin/go-cli-package watch {{args}} & cd docs && bun install && bun run dev

docs-build: docs-generate
	@echo "🏗️  Building documentation site..."
	cd docs && NODE_ENV=production bun run build

docs-preview:
	@echo "👀 Previewing built documentation..."
	cd docs && bun run preview

docs-clean:
	@echo "🧹 Cleaning documentation build artifacts..."
	rm -rf docs/dist docs/.astro docs/node_modules docs/src/content/docs/api

# Package distribution tasks
init-homebrew-tap:
	@echo "🍺 Initializing Homebrew tap repository..."
	./scripts/init_homebrew_tap.sh

init-aur-repo:
	@echo "📦 Initializing AUR repository..."
	./scripts/init_aur_repo.sh

update-homebrew-formula VERSION="":
	@echo "🍺 Updating Homebrew formula to version {{VERSION}}..."
	./scripts/update_homebrew_formula.sh {{VERSION}}

update-aur-pkgbuild VERSION="":
	@echo "📦 Updating AUR PKGBUILD..."
	./scripts/update_aur_pkgbuild.sh {{VERSION}}

# Git tag management
tag VERSION="":
	./scripts/tag.sh {{VERSION}}

tag-delete VERSION="":
	./scripts/tag_delete.sh {{VERSION}}

tag-list:
	@echo "📋 Available tags:"
	@git tag -l --sort=-v:refname | head -20

# Release management
release VERSION="":
	./scripts/release.sh {{VERSION}}

github-release VERSION="":
	./scripts/github_release.sh {{VERSION}}

deploy-aur VERSION="":
	./scripts/deploy_aur.sh {{VERSION}}

deploy-homebrew VERSION="":
	./scripts/deploy_homebrew.sh {{VERSION}}

deploy-all VERSION="":
	./scripts/deploy_all.sh {{VERSION}}

publish-homebrew VERSION="":
	./scripts/deploy_homebrew.sh {{VERSION}}

publish-aur VERSION="":
	./scripts/deploy_aur.sh {{VERSION}}

publish VERSION="":
	@just tag {{VERSION}}
	@just github-release {{VERSION}}
	@just release {{VERSION}}
	@just publish-homebrew {{VERSION}}
	@just publish-aur {{VERSION}}

