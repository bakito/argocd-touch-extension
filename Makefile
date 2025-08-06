# Include toolbox tasks
include ./.toolbox.mk

# Run go lint against code
lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

lint-ci: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: lint-ci
	go test -v ./...

release: tb.semver tb.goreleaser
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean --parallelism 2

test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean --parallelism 2

.PHONY: release
release: tb.semver tb.goreleaser
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean --parallelism 2

.PHONY: test-release
test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean --parallelism 2
