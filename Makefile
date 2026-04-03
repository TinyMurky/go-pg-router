BINARY_NAME=go-pg-router.out
MAIN_PATH=./cmd/go-pg-router


.PHONY: run
run: ## Run go-pg-router
	@go run $(MAIN_PATH)

.PHONY: build
build: ## Build go-pg-router
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: clean
clean: ## Clean binary file of go-pg-router
	@go clean
	@rm -f $(BINARY_NAME)

.PHONY: test
test: ## Run all tests with race detector
	@go test -race ./...

.PHONY: psql-test-conn-no-ssh
psql-test-conn-no-ssh: ## Run psql to test if it can connect to go-pg-router
	psql -h localhost -p 3000 -U anyuser -d "dbname=anydb sslmode=disable"

.PHONY:  help
help: ## Show help message
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## " } /^[a-zA-Z_-]+:.*?## / {printf "	\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
