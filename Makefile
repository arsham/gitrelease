help: ## Show help messages.
	@grep -E '^[0-9a-zA-Z_-]+:(.*?## .*)?$$' $(MAKEFILE_LIST) | sed 's/^Makefile://' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run="."
dir="./..."
short="-short"
flags=""
timeout=1m
build_tag=$(shell git describe --abbrev=0 --tags)
current_sha=$(shell git rev-parse --short HEAD)

.PHONY: install
install: ## Install gitrelease.
	@go install -trimpath -ldflags="-s -w -X main.version=$(build_tag) -X main.currentSha=$(current_sha)"

.PHONY: unit_test
unit_test: ## Run unit tests. You can set: [run, timeout, short, dir, flags]. Example: make unit_test flags="-race".
	@go mod tidy; go test -trimpath --timeout=$(timeout) $(short) $(dir) -run $(run) $(flags)

.PHONY: unit_test_watch
unit_test_watch: ## Run unit tests in watch mode. You can set: [run, timeout, short, dir, flags]. Example: make unit_test flags="-race".
	@echo "running tests on $(run). waiting for changes..."
	@-zsh -c "go mod tidy; go test -trimpath --timeout=$(timeout) $(short) $(dir) -run $(run) $(flags); repeat 100 printf '#'; echo"
	@reflex -d none -r "(\.go$$)|(go.mod)|(\.sql$$)" -- zsh -c "go mod tidy; go test -trimpath --timeout=$(timeout) $(short) $(dir) -run $(run) $(flags); repeat 100 printf '#'"

.PHONY: lint
lint: ## list the code
	go fmt ./...
	go vet ./...
	golangci-lint run ./...

.PHONY: ci_tests
ci_tests: ## Run tests for CI.
	go test -trimpath --timeout=10m -failfast -v -tags=integration -race -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: dependencies
dependencies: ## Install dependencies requried for development operations.
	@go install github.com/cespare/reflex@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2
	@go install github.com/psampaz/go-mod-outdated@latest
	@go install github.com/jondot/goweight@latest
	@go get -t -u golang.org/x/tools/cmd/cover
	@go get -t -u github.com/sonatype-nexus-community/nancy@latest
	@go get -u ./...
	@go mod tidy

.PHONY: run
run: ## Run the application. More like: make run args="update user"
	@go run . $(args)

.PHONY: clean
clean: ## Clean test caches and tidy up modules.
	@go clean -testcache
	@go mod tidy

.PHONY: coverage
coverage: ## Show the test coverage on browser.
	go test -covermode=count -coverprofile=coverage.out -tags=integration ./...
	go tool cover -func=coverage.out | tail -n 1
	go tool cover -html=coverage.out

.PHONY: audit
audit: ## Audit the code for updates, vulnerabilities and binary weight.
	go list -u -m -json all | go-mod-outdated -update -direct
	go list -json -m all | nancy sleuth
	goweight | head -n 20
