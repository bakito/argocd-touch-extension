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

.PHONY: release
release: tb.semver tb.goreleaser
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean --parallelism 2

.PHONY: test-release
test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean --parallelism 2


helm-docs: tb.helm-docs update-chart-version
	@$(TB_HELM_DOCS)

# Detect OS
OS := $(shell uname)

# Define the sed command based on OS
SED := $(if $(filter Darwin, $(OS)), sed -i "", sed -i)

update-chart-version: tb.semver
	@version=$$($(TB_SEMVER) -next); \
	versionNum=$$($(TB_SEMVER) -next -numeric); \
	$(SED) "s/^version:.*$$/version: $${versionNum}/"    ./helm/Chart.yaml; \
	$(SED) "s/^appVersion:.*$$/appVersion: $${version}/" ./helm/Chart.yaml

helm-lint: helm-docs
	helm lint ./helm

helm-template:
	helm template ./helm -n argo-cd -f testdata/test-values.yaml
