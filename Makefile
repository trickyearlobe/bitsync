APP    := bitsync
PKG    := .
BINDIR := bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT)

LATEST_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)

.DEFAULT_GOAL := help
.PHONY: help version build install test clean bump-major bump-minor bump-patch

help: ## List available targets
	@awk 'BEGIN { FS = ":.*?## " } /^[a-zA-Z_-]+:.*?## / { printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

version: ## Print the version that would be embedded in a build now
	@echo $(VERSION)

build: ## Build the binary into $(BINDIR)/
	@mkdir -p $(BINDIR)
	go build -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(APP) $(PKG)

install: ## Install to $$GOPATH/bin (or $$GOBIN)
	go install -ldflags "$(LDFLAGS)" $(PKG)

test: ## Run tests
	go test ./...

clean: ## Remove build artefacts
	rm -rf $(BINDIR)

bump-major: ## Tag the next major version (vX+1.0.0) — does not push
	@$(MAKE) -s _bump PART=major

bump-minor: ## Tag the next minor version (vX.Y+1.0) — does not push
	@$(MAKE) -s _bump PART=minor

bump-patch: ## Tag the next patch version (vX.Y.Z+1) — does not push
	@$(MAKE) -s _bump PART=patch

_bump:
	@current=$(LATEST_TAG); \
	stripped=$${current#v}; \
	major=$$(echo $$stripped | cut -d. -f1); \
	minor=$$(echo $$stripped | cut -d. -f2); \
	patch=$$(echo $$stripped | cut -d. -f3); \
	case "$(PART)" in \
		major) major=$$((major+1)); minor=0; patch=0 ;; \
		minor) minor=$$((minor+1)); patch=0 ;; \
		patch) patch=$$((patch+1)) ;; \
	esac; \
	new=v$$major.$$minor.$$patch; \
	echo "Tagging $$new (was $$current)"; \
	git tag -a $$new -m "Release $$new"; \
	echo; echo "Push it with: git push origin $$new"
