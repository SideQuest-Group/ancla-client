VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS  = -s -w \
           -X github.com/SideQuest-Group/ancla-client/internal/cli.Version=$(VERSION) \
           -X github.com/SideQuest-Group/ancla-client/internal/cli.Commit=$(COMMIT)

.PHONY: build install test vet fmt clean sync-openapi docs docs-dev docs-serve docs-gen

build: ## Build the ancla binary
	go build -ldflags '$(LDFLAGS)' -o dist/ancla ./cmd/ancla

install: build ## Build and install ancla to $GOBIN
	install dist/ancla $(shell go env GOBIN 2>/dev/null || echo $(shell go env GOPATH)/bin)/ancla

test: ## Run tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofmt -w .

clean: ## Remove build artifacts
	rm -rf dist/

sync-openapi: ## Copy openapi.json from the private ancla repo (set ANCLA_REPO)
	@if [ -z "$(ANCLA_REPO)" ]; then echo "Set ANCLA_REPO to the path of the ancla repo"; exit 1; fi
	cp "$(ANCLA_REPO)/openapi.json" openapi.json

docs-gen: ## Generate CLI + API reference docs
	go run ./cmd/gen-docs --out docs/src/content/docs/cli
	python3 scripts/gen-api-docs.py --spec openapi.json --out docs/src/content/docs/api

docs: docs-gen ## Build the documentation site
	cd docs && bun install && bun run build

docs-dev: docs-gen ## Start docs dev server on :4321
	cd docs && bun install && bun run dev

docs-serve: docs ## Serve built docs site for preview
	cd docs && bun run preview

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
