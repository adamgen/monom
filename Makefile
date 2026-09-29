.PHONY: help build test test-e2e test-cases fmt-check vet lint clean check site-build site-dev

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

# Runs every suite, then fails if any of them failed.
test-e2e: build ## Run shUnit2 e2e test suites
	@fail=0; for f in tests/mnmd_*_test tests/monom_*_test; do bash "$$f" || { echo "FAILED: $$f"; fail=1; }; done; exit $$fail

# Runs the harness module's tests: the YAML loader's unit tests and TestCases.
# -count=1: the cases run shells and bin/mnmd, which Go's test cache can't see.
# One file: CASES=tests/cases/monom_run.yaml make test-cases
test-cases: build ## Run the declarative CLI cases (tests/cases/*.yaml) in bash and zsh
	cd $(HARNESS) && go test -count=1 ./...

GO_DIRS = cmd internal $(HARNESS)

fmt-check: ## Fail if any Go file in either module needs gofmt
	@out=$$(gofmt -l $(GO_DIRS)); if [ -n "$$out" ]; then echo "gofmt -w needed:"; echo "$$out"; exit 1; fi

vet: ## Run go vet on both modules
	go vet ./...
	cd $(HARNESS) && go vet ./...

SHELL_FILES = tests/mnmd_*_test tests/monom_*_test tests/helpers src/monom src/monom.bash $(HARNESS)/testdata/tab.bash

lint: ## Run shellcheck on all shell files (zsh excluded: SC1071)
	shellcheck $(SHELL_FILES)

clean: ## Remove build artifacts
	rm -f bin/mnmd

check: build fmt-check vet ## Build, gofmt, vet, test, run e2e suites and cases, and lint
	go test ./...
	@$(MAKE) test-e2e
	@$(MAKE) test-cases
	shellcheck $(SHELL_FILES)

# The marketing site (site/) is not part of the runtime and not in `check`.
site-dev: ## Run the marketing site locally (needs Node 20+)
	cd site && npm install && npm run dev

site-build: ## Typecheck, build, and prerender the marketing site
	cd site && npm ci && npm run typecheck && npm run build && node scripts/prerender.mjs
