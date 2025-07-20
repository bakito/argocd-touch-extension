# Include toolbox tasks
include ./.toolbox.mk

# Run go lint against code
lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: lint
	go test -v ./...

release: tb.semver tb.goreleaser
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean --parallelism 2

test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean --parallelism 2
