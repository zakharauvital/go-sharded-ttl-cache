SHELL := /bin/bash
.DEFAULT_GOAL := help
.PHONY: *

COVERAGE_OUT_PATH := "./coverage.out"
COVERAGE_HTML_PATH := "./coverage.html"

help: ## Display available commands
	@printf "\033[0;36mAvailable commands:\033[0m\n"
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//'`); \
	for help_line in $${help_lines[@]}; do \
		IFS=$$'#' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//' -e 's/:.*//' -e 's/://'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf "\033[0;32m  %-15s \033[0m \t %s\n" $$help_command $$help_info ; \
	done

run: ## Run go run .
	@go run .

tidy: ## Run go mod tidy
	@go mod tidy

fmt: ## Run gofmt -w
	@printf "\033[0;36mRun gofmt...\033[0m\n"
	@gofmt -w .

lint: ## Run gofmt check and go vet
	@printf "\033[0;36mRun lint...\033[0m\n"
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
	@go vet ./...

test: ## Run go test -race
	@printf "\033[0;36mRun test...\033[0m\n"
	@go test -race -count=1 ./...

bench: ## Run go test -bench
	@printf "\033[0;36mRun bench...\033[0m\n"
	@go test -run=^$$ -bench=. -benchmem ./...

test-coverage: ## Run test coverage
	@go test -race -coverprofile=${COVERAGE_OUT_PATH} -covermode atomic ./internal/...
	@go tool cover -html=${COVERAGE_OUT_PATH} -o ${COVERAGE_HTML_PATH}
	@go tool cover -func ${COVERAGE_OUT_PATH} | grep total
	@echo "File is located in coverage.html"
