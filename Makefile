.PHONY: help test lint fmt vet build install run clean release-snapshot

BINARY := mira
PKG    := ./...

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## /{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run unit tests with race detector + coverage
	go test $(PKG) -race -coverprofile=coverage.out

lint: vet fmt-check ## Run vet + format check (and golangci-lint if installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed — skipping (CI runs it)"; \
	fi

vet: ## go vet
	go vet $(PKG)

fmt: ## Apply gofmt -s in place
	gofmt -s -w .

fmt-check: ## Fail if gofmt would change anything
	@unformatted=$$(gofmt -s -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt'd:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

build: ## Build the binary at the repo root
	go build -trimpath -o $(BINARY) .

install: ## go install into $$GOBIN
	go install .

run: ## Run from source against the current directory
	go run .

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out
	rm -rf dist/

release-snapshot: ## Dry-run a release with goreleaser (no publish)
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "goreleaser not installed — see https://goreleaser.com/install/"; exit 1; \
	fi
	goreleaser release --snapshot --clean
