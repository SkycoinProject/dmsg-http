.DEFAULT_GOAL := help
.PHONY : check lint install-linters dep test bin build

OPTS?=GO111MODULE=on GOBIN=${PWD}/bin
TEST_OPTS?=-race -tags no_ci -cover -timeout=5m
BIN_DIR?=./bin
BUILD_OPTS?=

check: lint test ## Run linters and tests

lint: ## Run linters. Use make install-linters first	
	${OPTS} golangci-lint run -c .golangci.yml ./...
	# The govet version in golangci-lint is out of date and has spurious warnings, run it separately
	${OPTS} go vet -all ./...

vendorcheck:  ## Run vendorcheck
	GO111MODULE=off vendorcheck ./...

test: ## Run tests
	-go clean -testcache &>/dev/null
	${OPTS} go test ${TEST_OPTS} ./...

install-linters: ## Install linters
	- VERSION=1.22.2 ./ci_scripts/install-golangci-lint.sh
	# GO111MODULE=off go get -u github.com/FiloSottile/vendorcheck
	# For some reason this install method is not recommended, see https://github.com/golangci/golangci-lint#install
	# However, they suggest `curl ... | bash` which we should not do
	# ${OPTS} go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	${OPTS} go get -u golang.org/x/tools/cmd/goimports

format: ## Formats the code. Must have goimports installed (use make install-linters).
	${OPTS} goimports -w -local github.com/SkycoinProject/dmsg .

dep: ## Sorts dependencies
	${OPTS} go mod download
	${OPTS} go mod tidy -v

build: ## Build binaries into ./bin
	${OPTS} go install ./cmd/*

bin: ## Build `dmsg-discovery`, `dmsg-server`
	${OPTS} go build ${BUILD_OPTS} -o ./dmsg-discovery ./cmd/dmsg-discovery
	${OPTS} go build ${BUILD_OPTS} -o ./dmsg-server  ./cmd/dmsg-server

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
