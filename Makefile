# Version info
VERSION ?= v0.1.6
GIT_COMMIT ?= $(shell git rev-parse HEAD)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Linker flags
LDFLAGS := -ldflags "-X github.com/forkspacer/api-server/pkg/version.Version=$(VERSION) -X github.com/forkspacer/api-server/pkg/version.GitCommit=$(GIT_COMMIT) -X github.com/forkspacer/api-server/pkg/version.BuildDate=$(BUILD_DATE)"

IMG ?= ghcr.io/forkspacer/api-server:$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
		$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
		$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
		$(GOLANGCI_LINT) config verify

.PHONY: dev
dev: fmt vet ## Run go vet against code.
	go run $(LDFLAGS) ./cmd/main.go

##@ Build

.PHONY: docker-build
docker-build: ## Build docker image
	$(CONTAINER_TOOL) build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image
	$(CONTAINER_TOOL) push ${IMG}

PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name api-builder
	$(CONTAINER_TOOL) buildx use api-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm api-builder
	rm Dockerfile.cross

##@ Version Management

.PHONY: update-version
update-version: ## Update version in Helm chart. Usage: make update-version VERSION=v1.0.1
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION not specified. Usage: make update-version VERSION=v1.0.1"; \
		exit 1; \
	fi
	@echo "🔄 Updating api-server to $(VERSION)..."
	@CHART_VERSION=$${VERSION#v}; \
	sed -i.bak "s/^version: .*/version: $$CHART_VERSION/" helm/Chart.yaml
	@sed -i.bak "s/^appVersion: .*/appVersion: \"$(VERSION)\"/" helm/Chart.yaml
	@sed -i.bak "s/tag: \".*\"/tag: \"$(VERSION)\"/" helm/values.yaml
	@sed -i.bak "s/^VERSION ?= .*/VERSION ?= $(VERSION)/" Makefile
	@rm -f helm/Chart.yaml.bak helm/values.yaml.bak Makefile.bak
	@echo "✅ Updated api-server to $(VERSION)"

##@ Helm Charts

.PHONY: helm-package
helm-package: ## Package Helm chart
	@echo "📦 Packaging Helm chart..."
	@CHART_VERSION=$$(grep '^version:' helm/Chart.yaml | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	helm package helm/
	@echo "✅ Helm chart packaged"

.PHONY: build-charts-site
build-charts-site: helm-package ## Create charts directory for GitHub Pages
	@CHART_VERSION=$$(grep '^version:' helm/Chart.yaml | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	CHART_FILE="api-server-$$CHART_VERSION.tgz"; \
	CHARTS_DIR="charts-site"; \
	echo "📦 Preparing charts site..."; \
	mkdir -p "$$CHARTS_DIR"
	@echo "📥 Fetching existing charts from GitHub Pages..."
	@if curl -fsSL https://forkspacer.github.io/api-server/index.yaml -o /tmp/current-index.yaml 2>/dev/null; then \
		echo "✅ Found existing Helm repository"; \
	else \
		echo "ℹ️  No existing charts found (first deployment)"; \
	fi
	@CHART_VERSION=$$(grep '^version:' helm/Chart.yaml | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	CHART_FILE="api-server-$$CHART_VERSION.tgz"; \
	CHARTS_DIR="charts-site"; \
	if [ -f /tmp/current-index.yaml ]; then \
		grep -oP 'https://forkspacer\.github\.io/api-server/api-server-[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?\.tgz' /tmp/current-index.yaml | sort -u | while read url; do \
			filename=$$(basename "$$url"); \
			if [ "$$filename" = "$$CHART_FILE" ]; then \
				echo "  ⏭️  Skipping $$filename (will be replaced)"; \
				continue; \
			fi; \
			echo "  📥 Downloading $$filename..."; \
			if curl -fsSL "$$url" -o "$$CHARTS_DIR/$$filename"; then \
				echo "  ✅ Downloaded $$filename"; \
			else \
				echo "  ⚠️  Failed to download $$filename"; \
			fi; \
		done; \
	fi
	@CHART_VERSION=$$(grep '^version:' helm/Chart.yaml | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	CHART_FILE="api-server-$$CHART_VERSION.tgz"; \
	CHARTS_DIR="charts-site"; \
	echo "✅ Downloaded $$(ls $$CHARTS_DIR/api-server-*.tgz 2>/dev/null | wc -l) existing chart(s)"; \
	cp "$$CHART_FILE" "$$CHARTS_DIR/"; \
	echo "✅ Added new chart: $$CHART_FILE"
	@CHARTS_DIR="charts-site"; \
	echo "📄 Generating Helm repo index..."; \
	helm repo index "$$CHARTS_DIR" --url https://forkspacer.github.io/api-server
	@CHARTS_DIR="charts-site"; \
	if [ -f ".github/templates/helm-page.html" ]; then \
		cp .github/templates/helm-page.html "$$CHARTS_DIR/index.html"; \
	else \
		echo "⚠️  Warning: .github/templates/helm-page.html not found, skipping HTML generation"; \
	fi
	@CHART_VERSION=$$(grep '^version:' helm/Chart.yaml | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	APP_VERSION=$$(grep '^appVersion:' helm/Chart.yaml | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	CHARTS_DIR="charts-site"; \
	if [ -f "$$CHARTS_DIR/index.html" ]; then \
		sed -i "s/{{VERSION_TAG}}/$$APP_VERSION/g" "$$CHARTS_DIR/index.html"; \
		sed -i "s/{{VERSION_NUMBER}}/$$CHART_VERSION/g" "$$CHARTS_DIR/index.html"; \
		echo "✅ Generated index.html from template"; \
	fi
	@CHARTS_DIR="charts-site"; \
	echo "✅ Charts site ready with $$(ls $$CHARTS_DIR/api-server-*.tgz 2>/dev/null | wc -l) chart version(s)"; \
	echo ""; \
	echo "📦 Available versions:"; \
	ls -lh $$CHARTS_DIR/api-server-*.tgz 2>/dev/null || echo "No charts found"

.PHONY: helm-summary
helm-summary: ## Generate GitHub Actions summary for Helm deployment
	@SUMMARY_FILE=$${GITHUB_STEP_SUMMARY:-/tmp/summary.md}; \
	CHART_VERSION=$$(grep '^version:' helm/Chart.yaml | awk '{print $$2}' | tr -d '"' | tr -d "'"); \
	CHART_FILE="api-server-$${CHART_VERSION}.tgz"; \
	echo "## 🎉 API Server Helm Charts Deployed" >> $$SUMMARY_FILE; \
	echo "- **Version**: $$CHART_VERSION" >> $$SUMMARY_FILE; \
	echo "- **Chart**: $$CHART_FILE" >> $$SUMMARY_FILE; \
	echo "- **Repository**: https://forkspacer.github.io/api-server" >> $$SUMMARY_FILE; \
	echo "" >> $$SUMMARY_FILE; \
	echo "### 📦 Installation" >> $$SUMMARY_FILE; \
	echo '```bash' >> $$SUMMARY_FILE; \
	echo "helm repo add forkspacer-api https://forkspacer.github.io/api-server" >> $$SUMMARY_FILE; \
	echo "helm repo update" >> $$SUMMARY_FILE; \
	echo "helm install api-server forkspacer-api/api-server" >> $$SUMMARY_FILE; \
	echo '```' >> $$SUMMARY_FILE; \
	echo "✅ Summary written to $$SUMMARY_FILE"

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

GOLANGCI_LINT_VERSION ?= v2.5.0

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $$(realpath $(1)-$(3)) $(1)
endef
