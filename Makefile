BIN = $(CURDIR)/bin

GOLANGCI_LINT_VERSION = v2.15.0
MOCKERY_VERSION = v3.9.1

GOLANGCI_LINT = $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
MOCKERY = $(BIN)/mockery-$(MOCKERY_VERSION)

.PHONY: check
check: lint test

.PHONY: lint
lint: $(GOLANGCI_LINT) mocks
	$(GOLANGCI_LINT) run

.PHONY: fix
fix: $(GOLANGCI_LINT) mocks
	$(GOLANGCI_LINT) run --fix

.PHONY: test
test: mocks
	go test --timeout 5m $(GO_TEST_FLAGS) ./...
	go test --timeout 5m $(GO_TEST_FLAGS) --race ./...
	go test --timeout 5m $(GO_TEST_FLAGS) --count 100 ./...

mocks: $(MOCKERY) .mockery.yml $(shell find . -type f -name '*.go' -not -name '*_test.go' -not -path './mocks/*')
	$(MOCKERY)
	@touch mocks

.PHONY: coverage
coverage: mocks
	go test --timeout 5m --covermode atomic --coverpkg . --coverprofile coverage.cov ./...

.PHONY: clean
clean:
	rm -rf bin coverage.cov

$(GOLANGCI_LINT):
	GOBIN=$(BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	mv $(BIN)/golangci-lint $@

$(MOCKERY):
	GOBIN=$(BIN) go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
	mv $(BIN)/mockery $@
