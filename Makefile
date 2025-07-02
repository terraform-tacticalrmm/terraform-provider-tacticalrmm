# Makefile for terraform-provider-trmm
# Systematic build automation for Terraform provider development

# Configuration Variables
HOSTNAME=registry.terraform.io
NAMESPACE=terraform-tacticalrmm
NAME=tacticalrmm
BINARY=terraform-provider-${NAME}
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

# Build Configuration
BUILD_FLAGS=-ldflags="-X 'main.version=${VERSION}'"
TEST_FLAGS=-v -cover -timeout=120s
ACCTEST_FLAGS=-v -timeout=120m

# Installation Paths
INSTALL_PATH=~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# Default target
.PHONY: default
default: build

# Build the provider binary
.PHONY: build
build:
	@echo "==> Building provider binary..."
	@echo "    Version: ${VERSION}"
	@echo "    Target:  ${OS_ARCH}"
	go build ${BUILD_FLAGS} -o ${BINARY}

# Install provider to local Terraform plugin directory
.PHONY: install
install: build
	@echo "==> Installing provider..."
	@echo "    Path: ${INSTALL_PATH}"
	@mkdir -p ${INSTALL_PATH}
	@cp ${BINARY} ${INSTALL_PATH}
	@echo "==> Installation complete"

# Run unit tests
.PHONY: test
test:
	@echo "==> Running unit tests..."
	go test ${TEST_FLAGS} ./...

# Run acceptance tests (requires TRMM instance)
.PHONY: testacc
testacc:
	@echo "==> Running acceptance tests..."
	@echo "    Note: Requires TRMM_ENDPOINT and TRMM_API_KEY environment variables"
	TF_ACC=1 go test ${ACCTEST_FLAGS} ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -f ${BINARY}
	@rm -rf dist/
	@echo "==> Clean complete"

# Format Go code
.PHONY: fmt
fmt:
	@echo "==> Formatting Go code..."
	@go fmt ./...
	@echo "==> Format complete"

# Run Go linters
.PHONY: lint
lint:
	@echo "==> Running linters..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
		exit 1; \
	fi

# Run static analysis
.PHONY: vet
vet:
	@echo "==> Running go vet..."
	@go vet ./...
	@echo "==> Vet complete"

# Generate documentation
.PHONY: docs
docs:
	@echo "==> Generating documentation..."
	@if command -v tfplugindocs >/dev/null; then \
		tfplugindocs generate; \
		echo "==> Documentation generated"; \
	else \
		echo "tfplugindocs not installed. Install with:"; \
		echo "  go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"; \
		exit 1; \
	fi

# Development build with race detection
.PHONY: dev
dev:
	@echo "==> Building development version with race detection..."
	go build -race ${BUILD_FLAGS} -o ${BINARY}

# Run all quality checks
.PHONY: check
check: fmt vet lint test
	@echo "==> All checks passed"

# Release preparation
.PHONY: release-prepare
release-prepare: clean check docs
	@echo "==> Release preparation complete"
	@echo "    Version: ${VERSION}"
	@echo "    Next steps:"
	@echo "      1. Commit all changes"
	@echo "      2. Tag the release: git tag vX.Y.Z"
	@echo "      3. Push tags: git push --tags"

# Cross-platform build
.PHONY: build-all
build-all:
	@echo "==> Building for all platforms..."
	@for platform in \
		"linux/amd64" \
		"linux/arm64" \
		"darwin/amd64" \
		"darwin/arm64" \
		"windows/amd64"; do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		go build ${BUILD_FLAGS} \
		-o dist/$${platform%/*}-$${platform#*/}/${BINARY} \
		&& echo "    Built: $${platform}"; \
	done
	@echo "==> Multi-platform build complete"

# Snapshot release (local testing)
.PHONY: snapshot
snapshot:
	@echo "==> Creating snapshot release..."
	@if command -v goreleaser >/dev/null; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed. Install with:"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi

# Update dependencies
.PHONY: deps
deps:
	@echo "==> Updating dependencies..."
	@go mod tidy
	@go mod verify
	@echo "==> Dependencies updated"

# Security scan
.PHONY: security
security:
	@echo "==> Running security scan..."
	@if command -v gosec >/dev/null; then \
		gosec -fmt=json -out=security-report.json ./... || true; \
		echo "==> Security scan complete (see security-report.json)"; \
	else \
		echo "gosec not installed. Install with:"; \
		echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate test coverage report
.PHONY: coverage
coverage:
	@echo "==> Generating test coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "==> Coverage report generated: coverage.html"

# Initialize development environment
.PHONY: init
init: deps
	@echo "==> Initializing development environment..."
	@go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/goreleaser/goreleaser@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "==> Development environment initialized"

# Help target
.PHONY: help
help:
	@echo "terraform-provider-trmm Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Build targets:"
	@echo "  build        Build the provider binary"
	@echo "  install      Build and install to local Terraform plugins"
	@echo "  dev          Build with race detection enabled"
	@echo "  build-all    Build for all supported platforms"
	@echo ""
	@echo "Test targets:"
	@echo "  test         Run unit tests"
	@echo "  testacc      Run acceptance tests (requires TRMM instance)"
	@echo "  coverage     Generate test coverage report"
	@echo ""
	@echo "Quality targets:"
	@echo "  fmt          Format Go code"
	@echo "  vet          Run go vet"
	@echo "  lint         Run linters"
	@echo "  security     Run security scan"
	@echo "  check        Run all quality checks"
	@echo ""
	@echo "Development targets:"
	@echo "  deps         Update dependencies"
	@echo "  docs         Generate documentation"
	@echo "  init         Initialize development environment"
	@echo ""
	@echo "Release targets:"
	@echo "  snapshot     Create local snapshot release"
	@echo "  release-prepare  Prepare for release"
	@echo ""
	@echo "Other targets:"
	@echo "  clean        Remove build artifacts"
	@echo "  help         Show this help message"
