SHELL:=/bin/bash

##@ Default Goal
.PHONY: help
help: ## Display this help
	@echo -e "Usage:\n  make <goal> [VAR=value ...]"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 } \
		/^##@/ { printf "\n%s\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development Goals

.PHONY: check
check: ## Run some code quality checks
	go vet ./...
	golint -set_exit_status ./...
	go test -race -short ./...

.PHONY: run
run: ## Run the service using "go run"
	go run ./main.go

.PHONY: build
build: check ## Build the executable
	CGO_ENABLED=0 go build -o epicctl ./main.go

.PHONY: install
install: check ## Install the exe to the go bin directory
	go install
