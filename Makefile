export BIN = ${PWD}/bin
export GOBIN = $(BIN)

.PHONY: check
check: lint test

.PHONY: lint
lint: $(BIN)/golangci-lint mocks
	$(BIN)/golangci-lint run

.PHONY: fix
fix: $(BIN)/golangci-lint mocks
	$(BIN)/golangci-lint run --fix

.PHONY: test
test: mocks
	go test --timeout 5m $(GO_TEST_FLAGS) ./...
	go test --timeout 5m $(GO_TEST_FLAGS) --race ./...
	go test --timeout 5m $(GO_TEST_FLAGS) --count 100 ./...

mocks: $(BIN)/mockery $(shell find . -type f -name '*.go' -not -name '*_test.go')
	$(BIN)/mockery --all --outpkg cmd_mocks
	@touch mocks

.PHONY: coverage
coverage: $(BIN)/go-acc mocks
	$(BIN)/go-acc --covermode set --output coverage.cov ./...

.PHONY: clean
clean:
	rm -rf bin mocks coverage.cov

$(BIN)/golangci-lint:
	curl --retry 5 -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh

$(BIN)/mockery:
	mkdir -p $(BIN)
	curl -sSfL https://api.github.com/repos/vektra/mockery/releases | jq -r 'first(.[] | select(.tag_name | test("^v2\\.8[.0-9]+$$"))) | .assets[] | select(.name | test("Linux_x86_64.tar.gz")) | .browser_download_url' \
		| xargs curl -sSfL \
		| tar -xz -C $(BIN) mockery
	#curl -sSfL https://api.github.com/repos/vektra/mockery/releases | jq -r 'first(.[] | select(.tag_name | test("^v2[.0-9]+$$"))) | .assets[] | select(.name | test("Linux_x86_64.tar.gz")) | .browser_download_url' \
#		| xargs curl -sSfL \
#		| tar -xz -C $(BIN) mockery

$(BIN)/go-acc:
	go install github.com/ory/go-acc@latest
