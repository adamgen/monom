.PHONY: help build test test-e2e lint clean check

help: ## Show available targets
	@awk -F'##' '/^[a-zA-Z_-]+[^#]*:.*##/ { split($$1, a, ":"); printf "  %-12s %s\n", a[1], $$2 }' $(MAKEFILE_LIST)

build: ## Compile bin/monomd
	@mkdir -p bin
	go build -o bin/monomd ./cmd/monomd

test: ## Run Go unit tests
	go test ./...

test-e2e: build ## Run shUnit2 e2e test suites
	@for f in tests/monomd_*_test; do bash "$$f"; done

lint: ## Run shellcheck on all shell files
	shellcheck tests/monomd_*_test tests/helpers
r
clean: ## Remove build artifacts
	rm -f bin/monomd

check: build ## Build, test, run e2e suites, and lint
	go test ./...
	@$(MAKE) test-e2e
	shellcheck tests/monomd_*_test tests/helpers
