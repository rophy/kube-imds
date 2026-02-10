.PHONY: help build build-client test test-e2e image

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build server binary to bin/kube-imds
	go build -o bin/kube-imds ./cmd/server

build-client: ## Build client binary to bin/kube-imds-client
	go build -o bin/kube-imds-client ./cmd/client

test: ## Run unit tests
	go test ./...

test-e2e: ## Run e2e tests (requires kind cluster)
	bats test/e2e/

image: ## Build Docker image via skaffold
	skaffold build
