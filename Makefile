VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS  = -s -w \
           -X github.com/SideQuest-Group/ancla-client/internal/cli.Version=$(VERSION) \
           -X github.com/SideQuest-Group/ancla-client/internal/cli.Commit=$(COMMIT)

.PHONY: build install test vet fmt fmt-check lint clean openapi docs docs-dev docs-serve docs-gen \
       spec-enrich sdk-go sdk-python sdk-typescript sdks openapi-full

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

fmt-check: ## Check formatting (CI)
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

lint: vet fmt-check ## Run all linting checks

clean: ## Remove build artifacts
	rm -rf dist/

ANCLA_REPO ?= ../ancla
OPENAPI_GEN_IMAGE ?= openapitools/openapi-generator-cli:v7.12.0
ENRICHED_SPEC     ?= openapi.enriched.json

openapi: ## Generate and pull openapi.json from the ancla backend repo
	$(MAKE) -C $(ANCLA_REPO) openapi
	cp $(ANCLA_REPO)/openapi.json openapi.json
	@echo "Updated openapi.json from $(ANCLA_REPO)"

spec-enrich: ## Enrich openapi.json with typed schemas and clean operationIds
	python3 scripts/enrich-openapi.py --spec openapi.json --out $(ENRICHED_SPEC)

sdk-go: spec-enrich ## Generate Go SDK from enriched spec
	cp codegen/openapi-generator-ignore-go sdks/go/.openapi-generator-ignore
	docker run --rm -v $(CURDIR):/work -w /work $(OPENAPI_GEN_IMAGE) generate -c codegen/go.yaml
	cd sdks/go && go mod tidy

sdk-python: spec-enrich ## Generate Python SDK from enriched spec
	cp codegen/openapi-generator-ignore-python sdks/python/.openapi-generator-ignore
	docker run --rm -v $(CURDIR):/work -w /work $(OPENAPI_GEN_IMAGE) generate -c codegen/python.yaml
	cd sdks/python && uv sync

sdk-typescript: spec-enrich ## Generate TypeScript SDK from enriched spec
	cp codegen/openapi-generator-ignore-ts sdks/typescript/.openapi-generator-ignore
	docker run --rm -v $(CURDIR):/work -w /work $(OPENAPI_GEN_IMAGE) generate -c codegen/typescript.yaml
	cd sdks/typescript && bun install

sdks: sdk-go sdk-python sdk-typescript ## Generate all SDKs

openapi-full: openapi sdks ## Pull fresh spec from backend and regenerate all SDKs

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
