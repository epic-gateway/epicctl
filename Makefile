SHELL:=/bin/bash

##@ Default Goal
.PHONY: help
help: ## Display this help
	@echo -e "Usage:\n  make <goal> [VAR=value ...]"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 } \
		/^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development Goals

.PHONY: clean
clean: ## Remove build artifacts
	rm -f docs/epicctl*.md epicctl

.PHONY: check
check: ## Run some code quality checks
	go vet ./...
	go test -race -short ./...

.PHONY: run
run: ## Run the service using "go run". Use the ARGS env var to pass params into go run.
	go run ./main.go ${ARGS}

.PHONY: build
build: check ## Build the executable
	CGO_ENABLED=0 go build -o epicctl .

.PHONY: install
install: check ## Install the exe to the go bin directory
	go install

.PHONY: docs
docs: ## Build documentation
	go run ./main.go markdown ./docs
