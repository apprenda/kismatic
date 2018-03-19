# Setup some useful vars
PKG = github.com/apprenda/kismatic
BUILD_OUTPUT = out-$(GOOS)

# Set the build version
ifeq ($(origin VERSION), undefined)
	VERSION := $(shell git describe --tags --always --dirty)
endif
# Set the build branch
ifeq ($(origin BRANCH), undefined)
	BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
endif
# build date
ifeq ($(origin BUILD_DATE), undefined)
	BUILD_DATE := $(shell date -u)
endif
# If no target is defined, assume the host is the target.
ifeq ($(origin GOOS), undefined)
	GOOS := $(shell go env GOOS)
endif
# Lots of these target goarches probably won't work,
# since we depend on vendored packages also being built for the correct arch
ifeq ($(origin GOARCH), undefined)
	GOARCH := $(shell go env GOARCH)
endif
# If no target is defined, assume the host is the target.
ifeq ($(origin HOST_GOOS), undefined)
	HOST_GOOS := $(shell go env GOOS)
endif
# Lots of these target goarches probably won't work,
# since we depend on vendored packages also being built for the correct arch
ifeq ($(origin HOST_GOARCH), undefined)
	HOST_GOARCH := $(shell go env GOARCH)
endif
# Used by the integration tests to tag nodes
ifeq ($(origin CREATED_BY), undefined)
	CREATED_BY := $(shell hostname)
endif

# Versions of external dependencies
GLIDE_VERSION = v0.13.1
ANSIBLE_VERSION = 2.3.0.0
PROVISIONER_VERSION = v1.9.0
KUBERANG_VERSION = v1.3.0
GO_VERSION = 1.9.4
KUBECTL_VERSION = v1.9.3
HELM_VERSION = v2.8.1

install: 
	@echo Building kismatic in container
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e HOST_GOOS="linux"                   \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                           \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w /go/src/$(PKG)                      \
	    circleci/golang:$(GO_VERSION)          \
	    make bin/$(GOOS)/kismatic build-inspector-host copy-kismatic copy-playbooks copy-inspector

dist: shallow-clean
	@echo "Running dist inside contianer"
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e HOST_GOOS="linux"                   \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                           \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w "/go/src/$(PKG)"                    \
	    circleci/golang:$(GO_VERSION)          \
	    make dist-common

clean: 
	rm -rf bin
	rm -rf out-*
	rm -rf vendor
	rm -rf vendor-*
	rm -rf tools
	rm -rf tmp

test:
	@docker run                             \
	    --rm                                \
	    -e HOST_GOOS="linux"                \
	    -u root:root                        \
	    -v "$(shell pwd)":/go/src/$(PKG)    \
	    -w /go/src/$(PKG)                   \
	    circleci/golang:$(GO_VERSION)       \
	    make test-host

integration-test: 
	mkdir -p tmp
	@echo "Running integration tests inside contianer"
	@docker run                                                 \
	    --rm                                                    \
	    -e GOOS="linux"                                         \
	    -e HOST_GOOS="linux"                                    \
	    -e VERSION="$(VERSION)"                                 \
	    -e BUILD_DATE="$(BUILD_DATE)"                           \
	    -e AWS_ACCESS_KEY_ID="$(AWS_ACCESS_KEY_ID)"             \
	    -e AWS_SECRET_ACCESS_KEY="$(AWS_SECRET_ACCESS_KEY)"     \
	    -e LEAVE_ARTIFACTS="$(LEAVE_ARTIFACTS)"                 \
	    -e CREATED_BY="$(CREATED_BY)"                           \
	    -u root:root                                            \
	    -v "$(shell pwd)":"/go/src/$(PKG)"                      \
	    -v "$(HOME)/.ssh/kismatic-integration-testing.pem":"/root/.ssh/kismatic-integration-testing.pem:ro" \
	    -v "$(shell pwd)/tmp":"/tmp/kismatic"                   \
	    -w "/go/src/$(PKG)"                                     \
	    circleci/golang:$(GO_VERSION)                           \
	    make integration-test-host

focus-integration-test: 
	mkdir -p tmp
	@echo "Running integration tests inside contianer"
	@docker run                                                 \
	    --rm                                                    \
	    -e FOCUS="$(FOCUS)"                                     \
	    -e GOOS="linux"                                         \
	    -e HOST_GOOS="linux"                                    \
	    -e VERSION="$(VERSION)"                                 \
	    -e BUILD_DATE="$(BUILD_DATE)"                           \
	    -e AWS_ACCESS_KEY_ID="$(AWS_ACCESS_KEY_ID)"             \
	    -e AWS_SECRET_ACCESS_KEY="$(AWS_SECRET_ACCESS_KEY)"     \
	    -e LEAVE_ARTIFACTS="$(LEAVE_ARTIFACTS)"                 \
	    -e CREATED_BY="$(CREATED_BY)"                           \
	    -u root:root                                            \
	    -v "$(shell pwd)":"/go/src/$(PKG)"                      \
	    -v "$(HOME)/.ssh/kismatic-integration-testing.pem":"/root/.ssh/kismatic-integration-testing.pem:ro" \
	    -v "$(shell pwd)/tmp":"/tmp/kismatic"                   \
	    -w "/go/src/$(PKG)"                                     \
	    circleci/golang:$(GO_VERSION)                           \
	    make focus-integration-test-host

slow-integration-test: 
	mkdir -p tmp
	@echo "Running integration tests inside contianer"
	@docker run                                                 \
	    --rm                                                    \
	    -e GOOS="linux"                                         \
	    -e HOST_GOOS="linux"                                    \
	    -e VERSION="$(VERSION)"                                 \
	    -e BUILD_DATE="$(BUILD_DATE)"                           \
	    -e AWS_ACCESS_KEY_ID="$(AWS_ACCESS_KEY_ID)"             \
	    -e AWS_SECRET_ACCESS_KEY="$(AWS_SECRET_ACCESS_KEY)"     \
	    -e LEAVE_ARTIFACTS="$(LEAVE_ARTIFACTS)"                 \
	    -e CREATED_BY="$(CREATED_BY)"                           \
	    -u root:root                                            \
	    -v "$(shell pwd)":"/go/src/$(PKG)"                      \
	    -v "$(HOME)/.ssh/kismatic-integration-testing.pem":"/root/.ssh/kismatic-integration-testing.pem:ro" \
	    -v "$(shell pwd)/tmp":"/tmp/kismatic"                   \
	    -w "/go/src/$(PKG)"                                     \
	    circleci/golang:$(GO_VERSION)                           \
	    make slow-integration-test-host

# YOU SHOULDN'T NEED TO USE ANYTHING BENEATH THIS LINE
# UNLESS YOU REALLY KNOW WHAT YOU'RE DOING
# ---------------------------------------------------------------------
all:
	@$(MAKE) GOOS=darwin dist
	@$(MAKE) GOOS=linux dist

shallow-clean:
	rm -rf $(BUILD_OUTPUT)

tar-clean: 
	rm kismatic-*.tar.gz

build: 
	@echo Building kismatic in container
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e HOST_GOOS="linux"                   \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                           \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w /go/src/$(PKG)                      \
	    circleci/golang:$(GO_VERSION)          \
	    make build-host

build-inspector:
	@echo Building inspector in container
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e HOST_GOOS="linux"                   \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                           \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w /go/src/$(PKG)                      \
	    circleci/golang:$(GO_VERSION)          \
	    make build-inspector-host

glide-install:
	@echo Glide installing in container
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e HOST_GOOS="linux"                   \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                           \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w /go/src/$(PKG)                      \
	    circleci/golang:$(GO_VERSION)          \
	    make glide-install-host

glide-update:
	@echo Glide updating in container
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e HOST_GOOS="linux"                   \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                           \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w /go/src/$(PKG)                      \
	    circleci/golang:$(GO_VERSION)          \
	    make glide-update-host

copy-all: copy-vendors copy-inspector copy-playbooks copy-kismatic

copy-kismatic:
	mkdir -p $(BUILD_OUTPUT)
	cp bin/$(GOOS)/kismatic $(BUILD_OUTPUT)

copy-inspector:
	rm -rf $(BUILD_OUTPUT)/ansible/playbooks/inspector
	mkdir -p $(BUILD_OUTPUT)/ansible/playbooks/inspector
	cp -r bin/inspector/* $(BUILD_OUTPUT)/ansible/playbooks/inspector

copy-playbooks:
	mkdir -p $(BUILD_OUTPUT)/ansible
	rm -rf $(filter-out $(BUILD_OUTPUT)/ansible/playbooks/inspector $(BUILD_OUTPUT)/ansible/playbooks/kuberang, $(wildcard $(BUILD_OUTPUT)/ansible/playbooks/*))
	cp -r $(wildcard ansible/*) $(BUILD_OUTPUT)/ansible/playbooks

copy-vendors: # omit kismatic, inspector, terraform since we provide configs for those.
	mkdir -p $(BUILD_OUTPUT)/ansible
	cp -r vendor-ansible/out/ansible/* $(BUILD_OUTPUT)/ansible
	cp vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-$(GOARCH) $(BUILD_OUTPUT)/kubectl
	cp vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH) $(BUILD_OUTPUT)/helm
	cp vendor-provision/out/provision-$(PROVISIONER_VERSION)-$(GOOS)-$(GOARCH) $(BUILD_OUTPUT)/provision
	mkdir -p $(BUILD_OUTPUT)/ansible/playbooks/kuberang/linux/$(GOARCH)/
	cp vendor-kuberang/$(KUBERANG_VERSION)/kuberang-linux-$(GOARCH) $(BUILD_OUTPUT)/ansible/playbooks/kuberang/linux/$(GOARCH)/kuberang

tarball: 
	rm -f kismatic-$(GOOS).tar.gz
	tar -czf kismatic-$(GOOS).tar.gz -C $(BUILD_OUTPUT) .

# RECIPES BELOW THIS LINE ARE INTENDED FOR CI ONLY. RUN LOCALLY AT YOUR OWN RISK.
# ---------------------------------------------------------------------

all-host:
	@$(MAKE) GOOS=darwin dist-host
	@$(MAKE) GOOS=linux dist-host

test-host:
	go test ./cmd/... ./pkg/... $(TEST_OPTS)

build-host: tools/glide-$(HOST_GOOS)-$(HOST_GOARCH) glide-install-host bin/$(GOOS)/kismatic

.PHONY: bin/$(GOOS)/kismatic
bin/$(GOOS)/kismatic:
	go build -o $@                                                              \
	    -ldflags "-X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'" \
	    ./cmd/kismatic

build-inspector-host:
	@$(MAKE) GOOS=linux bin/inspector/linux/$(GOARCH)/kismatic-inspector

.PHONY: bin/inspector/$(GOOS)/$(GOARCH)/kismatic-inspector
bin/inspector/$(GOOS)/$(GOARCH)/kismatic-inspector:
	go build -o $@                                                               \
	    -ldflags "-X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'"  \
	    ./cmd/kismatic-inspector

glide-install-host:
	tools/glide-$(HOST_GOOS)-$(HOST_GOARCH) cc
	tools/glide-$(HOST_GOOS)-$(HOST_GOARCH) install

glide-update-host:
	tools/glide-$(HOST_GOOS)-$(HOST_GOARCH) update

vendor: vendor-tools vendor-ansible/out vendor-provision/out/provision-$(PROVISIONER_VERSION)-$(GOOS)-$(GOARCH) vendor-kuberang/$(KUBERANG_VERSION) vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-$(GOARCH) vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH)

vendor-tools: tools/glide-$(HOST_GOOS)-$(HOST_GOARCH)

tools/glide-$(HOST_GOOS)-$(HOST_GOARCH):
	mkdir -p tools
	curl -L https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(GLIDE_VERSION)-$(HOST_GOOS)-$(HOST_GOARCH).tar.gz | tar -xz -C tools
	mv tools/$(HOST_GOOS)-$(HOST_GOARCH)/glide tools/glide-$(HOST_GOOS)-$(HOST_GOARCH)
	rm -r tools/$(HOST_GOOS)-$(HOST_GOARCH)

vendor-ansible/out:
	mkdir -p vendor-ansible/out
	curl -L https://github.com/apprenda/vendor-ansible/releases/download/v$(ANSIBLE_VERSION)/ansible.tar.gz -o vendor-ansible/out/ansible.tar.gz
	tar -zxf vendor-ansible/out/ansible.tar.gz -C vendor-ansible/out
	rm vendor-ansible/out/ansible.tar.gz

vendor-provision/out/provision-$(PROVISIONER_VERSION)-$(GOOS)-$(GOARCH):
	mkdir -p vendor-provision/out/
	curl -L https://github.com/apprenda/kismatic-provision/releases/download/$(PROVISIONER_VERSION)/provision-$(GOOS)-$(GOARCH) -o vendor-provision/out/provision-$(PROVISIONER_VERSION)-$(GOOS)-$(GOARCH)
	chmod +x vendor-provision/out/provision-$(PROVISIONER_VERSION)-$(GOOS)-$(GOARCH)

vendor-kuberang/$(KUBERANG_VERSION):
	mkdir -p vendor-kuberang/$(KUBERANG_VERSION)
	curl -L https://github.com/apprenda/kuberang/releases/download/$(KUBERANG_VERSION)/kuberang-linux-$(GOARCH) -o vendor-kuberang/$(KUBERANG_VERSION)/kuberang-linux-$(GOARCH)

vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-$(GOARCH):
	mkdir -p vendor-kubectl/out/
	curl -L https://storage.googleapis.com/kubernetes-release/release/$(KUBECTL_VERSION)/bin/$(GOOS)/$(GOARCH)/kubectl -o vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-$(GOARCH)
	chmod +x vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-$(GOARCH)

vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH):
	mkdir -p vendor-helm/out/
	curl -L https://storage.googleapis.com/kubernetes-helm/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH).tar.gz | tar zx -C vendor-helm
	cp vendor-helm/$(GOOS)-$(GOARCH)/helm vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH)
	rm -rf vendor-helm/$(GOOS)-$(GOARCH)
	chmod +x vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-$(GOARCH)

dist-common: vendor build-host build-inspector-host copy-all

dist-host: shallow-clean dist-common

get-ginkgo:
	go get github.com/onsi/ginkgo/ginkgo
	cd integration-tests

integration-test-host: get-ginkgo
	@$(MAKE) GOOS=linux tarball
	ginkgo --skip "\[slow\]" -p $(GINKGO_OPTS) -v integration-tests

slow-integration-test-host: get-ginkgo
	@$(MAKE) GOOS=linux tarball
	ginkgo --focus "\[slow\]" -p $(GINKGO_OPTS) -v integration-tests

focus-integration-test-host: get-ginkgo
	@$(MAKE) GOOS=linux tarball
	ginkgo --focus $(FOCUS) $(GINKGO_OPTS) -v integration-tests

docs/generate-kismatic-cli:
	mkdir -p docs/kismatic-cli
	go run cmd/kismatic-docs/main.go
	cp docs/kismatic-cli/kismatic.md docs/kismatic-cli/README.md

docs/update-plan-file-reference.md:
	@$(MAKE) docs/generate-plan-file-reference.md > docs/plan-file-reference.md

docs/generate-plan-file-reference.md:
	@go run cmd/gen-kismatic-ref-docs/*.go -o markdown pkg/install/plan_types.go Plan

version:
	@echo VERSION=$(VERSION)
	@echo GLIDE_VERSION=$(GLIDE_VERSION)
	@echo ANSIBLE_VERSION=$(ANSIBLE_VERSION)
	@echo PROVISIONER_VERSION=$(PROVISIONER_VERSION)

CIRCLE_ENDPOINT=
ifndef CIRCLE_CI_BRANCH
	CIRCLE_ENDPOINT=https://circleci.com/api/v1.1/project/github/apprenda/kismatic
else
	CIRCLE_ENDPOINT=https://circleci.com/api/v1.1/project/github/apprenda/kismatic/tree/$(CIRCLE_CI_BRANCH)
endif

trigger-ci-slow-tests:
	@echo Triggering build with slow tests
	curl -u $(CIRCLE_CI_TOKEN): -X POST --header "Content-Type: application/json"    \
	  -d '{"build_parameters": {"RUN_SLOW_TESTS": "true"}}'                          \
	  $(CIRCLE_ENDPOINT)
trigger-ci-focused-tests:
	@echo Triggering focused test
	curl -u $(CIRCLE_CI_TOKEN): -X POST --header "Content-Type: application/json"    \
	  -d "{\"build_parameters\": {\"FOCUS\": \"$(FOCUS)\"}}"                         \
	  $(CIRCLE_ENDPOINT)
