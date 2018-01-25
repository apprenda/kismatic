# Set the build version
ifeq ($(origin VERSION), undefined)
	VERSION := $(shell git describe --tags --always --dirty)
endif
# build date
ifeq ($(origin BUILD_DATE), undefined)
	BUILD_DATE := $(shell date -u)
endif

# Setup some useful vars
PKG = github.com/apprenda/kismatic
HOST_GOOS = $(shell go env GOOS)
HOST_GOARCH = $(shell go env GOARCH)

# Versions of external dependencies
GLIDE_VERSION = v0.13.1
ANSIBLE_VERSION = 2.3.0.0
PROVISIONER_VERSION = v1.7.0
KUBERANG_VERSION = v1.2.2
GO_VERSION = 1.9.2
KUBECTL_VERSION = v1.9.2
HELM_VERSION = v2.7.2

ifeq ($(origin GLIDE_GOOS), undefined)
	GLIDE_GOOS := $(HOST_GOOS)
endif
ifeq ($(origin GOOS), undefined)
	GOOS := $(HOST_GOOS)
endif

build: vendor # vendor on host because of some permission issues with glide inside container
	@echo Building kismatic in container
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e GLIDE_GOOS="linux"                  \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                 \
	    -v "$(shell pwd)":"/go/src/$(PKG)"      \
	    -w /go/src/$(PKG)                      \
	    circleci/golang:$(GO_VERSION)          \
	    make bare-build

bare-build: bin/$(GOOS)/kismatic

bare-build-update-dist: bare-build
	cp bin/$(GOOS)/kismatic out

build-inspector: vendor
	@echo Building inspector in container
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e GLIDE_GOOS="linux"                  \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                 \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w /go/src/$(PKG)                      \
	    circleci/golang:$(GO_VERSION)          \
	    make bare-build-inspector

bare-build-inspector: vendor
	@$(MAKE) GOOS=linux bin/inspector/linux/amd64/kismatic-inspector
	@$(MAKE) GOOS=darwin bin/inspector/darwin/amd64/kismatic-inspector

.PHONY: bin/$(GOOS)/kismatic
bin/$(GOOS)/kismatic: vendor
	go build -o $@                                                              \
	    -ldflags "-X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'" \
	    ./cmd/kismatic

.PHONY: bin/inspector/$(GOOS)/amd64/kismatic-inspector
bin/inspector/$(GOOS)/amd64/kismatic-inspector: vendor
	go build -o $@                                                               \
	    -ldflags "-X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'"  \
	    ./cmd/kismatic-inspector

clean:
	rm -rf bin
	rm -rf out
	rm -rf vendor
	rm -rf vendor-ansible
	rm -rf vendor-provision
	rm -rf integration-tests/vendor
	rm -rf vendor-kuberang
	rm -rf vendor-helm
	rm -rf vendor-kubectl
	rm -rf tools

test: vendor
	@docker run                             \
	    --rm                                \
	    -e GLIDE_GOOS="linux"               \
	    -u root:root              \
	    -v "$(shell pwd)":/go/src/$(PKG)    \
	    -v /tmp:/tmp                        \
	    -w /go/src/$(PKG)                   \
	    circleci/golang:$(GO_VERSION)       \
	    make bare-test

bare-test: vendor
	go test ./cmd/... ./pkg/... $(TEST_OPTS)

integration-test: dist just-integration-test

.PHONY: vendor
vendor: tools/glide-$(GLIDE_GOOS)-$(HOST_GOARCH)
	tools/glide-$(GLIDE_GOOS)-$(HOST_GOARCH) install

tools/glide-$(GLIDE_GOOS)-$(HOST_GOARCH):
	mkdir -p tools
	curl -L https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(GLIDE_VERSION)-$(GLIDE_GOOS)-$(HOST_GOARCH).tar.gz | tar -xz -C tools
	mv tools/$(GLIDE_GOOS)-$(HOST_GOARCH)/glide tools/glide-$(GLIDE_GOOS)-$(HOST_GOARCH)
	rm -r tools/$(GLIDE_GOOS)-$(HOST_GOARCH)

vendor-ansible/out:
	mkdir -p vendor-ansible/out
	curl -L https://github.com/apprenda/vendor-ansible/releases/download/v$(ANSIBLE_VERSION)/ansible.tar.gz -o vendor-ansible/out/ansible.tar.gz
	tar -zxf vendor-ansible/out/ansible.tar.gz -C vendor-ansible/out
	rm vendor-ansible/out/ansible.tar.gz

vendor-provision/out:
	mkdir -p vendor-provision/out/
	curl -L https://github.com/apprenda/kismatic-provision/releases/download/$(PROVISIONER_VERSION)/provision-darwin-amd64 -o vendor-provision/out/provision-darwin-amd64
	curl -L https://github.com/apprenda/kismatic-provision/releases/download/$(PROVISIONER_VERSION)/provision-linux-amd64 -o vendor-provision/out/provision-linux-amd64
	chmod +x vendor-provision/out/*

vendor-kuberang/$(KUBERANG_VERSION):
	mkdir -p vendor-kuberang/$(KUBERANG_VERSION)
	curl -L https://github.com/apprenda/kuberang/releases/download/$(KUBERANG_VERSION)/kuberang-linux-amd64 -o vendor-kuberang/$(KUBERANG_VERSION)/kuberang-linux-amd64

vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-amd64:
	mkdir -p vendor-kubectl/out/
	curl -L https://storage.googleapis.com/kubernetes-release/release/$(KUBECTL_VERSION)/bin/$(GOOS)/amd64/kubectl -o vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-amd64
	chmod +x vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-amd64

vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-amd64:
	mkdir -p vendor-helm/out/
	curl -L https://storage.googleapis.com/kubernetes-helm/helm-$(HELM_VERSION)-$(GOOS)-amd64.tar.gz | tar zx -C vendor-helm
	cp vendor-helm/$(GOOS)-amd64/helm vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-amd64
	rm -rf vendor-helm/$(GOOS)-amd64
	chmod +x vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-amd64

dist: vendor
	@echo "Running dist inside contianer"
	@docker run                                \
	    --rm                                   \
	    -e GOOS="$(GOOS)"                      \
	    -e GLIDE_GOOS="linux"                  \
	    -e VERSION="$(VERSION)"                \
	    -e BUILD_DATE="$(BUILD_DATE)"          \
	    -u root:root                 \
	    -v "$(shell pwd)":"/go/src/$(PKG)"     \
	    -w "/go/src/$(PKG)"                    \
	    circleci/golang:$(GO_VERSION)          \
	    make bare-dist

bare-dist: vendor-ansible/out vendor-provision/out vendor-kuberang/$(KUBERANG_VERSION) vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-amd64 vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-amd64 bare-build bare-build-inspector
	mkdir -p out
	cp bin/$(GOOS)/kismatic out
	mkdir -p out/ansible
	cp -r vendor-ansible/out/ansible/* out/ansible
	rm -rf out/ansible/playbooks
	cp -r ansible out/ansible/playbooks
	mkdir -p out/ansible/playbooks/inspector
	cp -r bin/inspector/* out/ansible/playbooks/inspector
	mkdir -p out/ansible/playbooks/kuberang/linux/amd64/
	cp vendor-kuberang/$(KUBERANG_VERSION)/kuberang-linux-amd64 out/ansible/playbooks/kuberang/linux/amd64/kuberang
	cp vendor-provision/out/provision-$(GOOS)-amd64 out/provision
	cp vendor-kubectl/out/kubectl-$(KUBECTL_VERSION)-$(GOOS)-amd64 out/kubectl
	cp vendor-helm/out/helm-$(HELM_VERSION)-$(GOOS)-amd64 out/helm
	rm -f out/kismatic.tar.gz
	tar -czf kismatic.tar.gz -C out .
	mv kismatic.tar.gz out

integration-tests/vendor: tools/glide-$(GLIDE_GOOS)-$(HOST_GOARCH)
	go get github.com/onsi/ginkgo/ginkgo
	cd integration-tests && ../tools/glide-$(GLIDE_GOOS)-$(HOST_GOARCH) install

just-integration-test: integration-tests/vendor
	ginkgo --skip "\[slow\]" -p $(GINKGO_OPTS) -v integration-tests

slow-integration-test: integration-tests/vendor
	ginkgo --focus "\[slow\]" -p $(GINKGO_OPTS) -v integration-tests

serial-integration-test: integration-tests/vendor
	ginkgo -v integration-tests

focus-integration-test: integration-tests/vendor
	ginkgo --focus $(FOCUS) $(GINKGO_OPTS) -v integration-tests

docs/generate-kismatic-cli:
	mkdir -p docs/kismatic-cli
	go run cmd/kismatic-docs/main.go
	cp docs/kismatic-cli/kismatic.md docs/kismatic-cli/README.md

docs/update-plan-file-reference.md:
	@$(MAKE) docs/generate-plan-file-reference.md > docs/plan-file-reference.md

docs/generate-plan-file-reference.md:
	@go run cmd/gen-kismatic-ref-docs/*.go -o markdown pkg/install/plan_types.go Plan

version: FORCE
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
	curl -u $(CIRCLE_CI_TOKEN): -X POST --header "Content-Type: application/json"     \
		-d '{"build_parameters": {"RUN_SLOW_TESTS": "true"}}'                         \
		$(CIRCLE_ENDPOINT)
trigger-ci-focused-tests:
	@echo Triggering focused test
	curl -u $(CIRCLE_CI_TOKEN): -X POST --header "Content-Type: application/json"     \
		-d "{\"build_parameters\": {\"FOCUS\": \"$(FOCUS)\"}}"                         \
		$(CIRCLE_ENDPOINT)

FORCE:
