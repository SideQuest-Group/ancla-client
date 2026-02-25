VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS  = -s -w \
           -X github.com/SideQuest-Group/ancla-client/internal/cli.Version=$(VERSION) \
           -X github.com/SideQuest-Group/ancla-client/internal/cli.Commit=$(COMMIT)

.PHONY: build test vet fmt clean sync-openapi

build: ## Build the ancla binary
	go build -ldflags '$(LDFLAGS)' -o dist/ancla ./cmd/ancla

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

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
