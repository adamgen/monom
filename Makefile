.PHONY: help build test test-e2e test-cases lint clean check

# The case runner is a separate Go module (tests/harness). Workspaces are off
# so a local go.work can't change how either module resolves dependencies.
export GOWORK := off
HARNESS := tests/harness

help: ## Show available targets
	@awk -F'##' '/^[a-zA-Z_-]+[^#]*:.*##/ { split($$1, a, ":"); printf "  %-12s %s\n", a[1], $$2 }' $(MAKEFILE_LIST)

build: ## Compile bin/mnmd
	@mkdir -p bin
	go build -o bin/mnmd ./cmd/mnmd

test: ## Run Go unit tests (root module; the tests/harness module is separate)
	go test ./...

test-e2e: build ## Run shUnit2 e2e test suites
	@for f in tests/mnmd_*_test tests/monom_*_test; do bash "$$f"; done

# Runs the harness module's tests: the YAML loader's unit tests and TestCases.
# -count=1: the cases run shells and bin/mnmd, which Go's test cache can't see.
# One file: CASES=tests/cases/monom_run.yaml make test-cases
test-cases: build ## Run the declarative CLI cases (tests/cases/*.yaml) in bash and zsh
	cd $(HARNESS) && go test -count=1 ./...

SHELL_FILES = tests/mnmd_*_test tests/monom_*_test tests/helpers src/monom src/monom.bash $(HARNESS)/testdata/tab.bash

lint: ## Run shellcheck on all shell files (zsh excluded: SC1071)
	shellcheck $(SHELL_FILES)

clean: ## Remove build artifacts
	rm -f bin/mnmd

check: build ## Build, vet, test, run e2e suites and cases, and lint
	go vet ./...
	cd $(HARNESS) && go vet ./...
	go test ./...
	@$(MAKE) test-e2e
	@$(MAKE) test-cases
	shellcheck $(SHELL_FILES)
