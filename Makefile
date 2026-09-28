.PHONY: help build test test-e2e test-cases lint clean check

help: ## Show available targets
	@awk -F'##' '/^[a-zA-Z_-]+[^#]*:.*##/ { split($$1, a, ":"); printf "  %-12s %s\n", a[1], $$2 }' $(MAKEFILE_LIST)

build: ## Compile bin/mnmd
	@mkdir -p bin
	go build -o bin/mnmd ./cmd/mnmd

test: ## Run Go unit tests (the cases runner is excluded: it needs -tags cases)
	go test ./...

test-e2e: build ## Run shUnit2 e2e test suites
	@for f in tests/mnmd_*_test tests/monom_*_test; do bash "$$f"; done

# -count=1: the cases run shells and bin/mnmd, which Go's test cache can't see.
# One file: CASES=tests/cases/monom_run.yaml make test-cases
test-cases: build ## Run the declarative CLI cases (tests/cases/*.yaml) in bash and zsh
	go test -tags cases -count=1 -run '^TestCases$$' ./internal/testcases

SHELL_FILES = tests/mnmd_*_test tests/monom_*_test tests/helpers src/monom src/monom.bash internal/testcases/testdata/tab.bash

lint: ## Run shellcheck on all shell files (zsh excluded: SC1071)
	shellcheck $(SHELL_FILES)

clean: ## Remove build artifacts
	rm -f bin/mnmd

check: build ## Build, vet, test, run e2e suites and cases, and lint
	go vet -tags cases ./...
	go test ./...
	@$(MAKE) test-e2e
	@$(MAKE) test-cases
	shellcheck $(SHELL_FILES)
