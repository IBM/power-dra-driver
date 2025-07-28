# This project applies to ppc64le only
ARCH ?= "ppc64le"

REGISTRY ?= quay.io/powercloud
REPOSITORY ?= power-dra-driver
TAG ?= v0.1.0

CONTAINER_RUNTIME ?= $(shell command -v podman 2> /dev/null || echo docker)

GOLANG_VERSION ?= 1.23.1

DRIVER_NAME := power-dra-driver
MODULE := github.com/IBM/$(DRIVER_NAME)

REGISTRY ?= quay.io/powercloud/power-dra-driver

VERSION ?= v0.1.0

GIT_COMMIT ?= $(shell git describe --match="" --dirty --long --always --abbrev=40 2> /dev/null || echo "")

# Kind configuration
ifeq ($(ARCH),ppc64le)
  KIND_IMAGE := quay.io/powercloud/kind-node:v1.33.1
else
  KIND_IMAGE := docker.io/kindest/node:latest
endif

KIND_CLUSTER_NAME:="power-dra-driver-cluster"
KIND_CLUSTER_CONFIG_PATH:="hack/kind-cluster-config.yaml"
KIND_EXPERIMENTAL_PROVIDER:="podman"

########################################################################
# Go Targets
.PHONY: build
build: fmt vet
	GOOS=linux GOARCH=$(ARCH) go build -o bin/power-dra-kubeletplugin cmd/power-dra-kubeletplugin/*.go

CONTROLLER_GEN := $(shell which controller-gen 2>/dev/null || echo "$$(go env GOPATH)/bin/controller-gen")

.PHONY: controller-gen
controller-gen:
    @if [ ! -x "$(CONTROLLER_GEN)" ]; then \
        echo "Installing controller-gen..."; \
        go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0; \
    fi

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -f ./bin/power-dra-driver
	rm -rf vendor

BUILDIMAGE_TAG ?= golang$(GOLANG_VERSION)
BUILDIMAGE ?= $(IMAGE_NAME)-build:$(BUILDIMAGE_TAG)

CMDS := $(patsubst ./cmd/%/,%,$(sort $(dir $(wildcard ./cmd/*/))))
CMD_TARGETS := $(patsubst %,cmd-%, $(CMDS))

GOOS ?= linux
GOARCH ?= ppc64le
ifeq ($(VERSION),)
CLI_VERSION = $(LIB_VERSION)$(if $(LIB_TAG),-$(LIB_TAG))
else
CLI_VERSION = $(VERSION)
endif
CLI_VERSION_PACKAGE = $(MODULE)/internal/info

CMDS := $(patsubst ./cmd/%/,%,$(sort $(dir $(wildcard ./cmd/*/))))
CMD_TARGETS := $(patsubst %,cmd-%, $(CMDS))

.PHONY: binaries
binaries: cmds
ifneq ($(PREFIX),)
cmd-%: COMMAND_BUILD_OPTIONS = -o $(PREFIX)/$(*)
endif

.PHONY: cmds
cmds: $(CMD_TARGETS)
$(CMD_TARGETS): cmd-%:
	CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' \
		CC=$(CC) CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "-s -w -X $(CLI_VERSION_PACKAGE).gitCommit=$(GIT_COMMIT) -X $(CLI_VERSION_PACKAGE).version=$(VERSION)" $(COMMAND_BUILD_OPTIONS) $(MODULE)/cmd/$(*)

########################################################################
# Testing Targets

.PHONY: dev-install-kind
dev-install-kind:
	mkdir -p dev-cache
	GOBIN=$(PWD)/dev-cache/ go install sigs.k8s.io/kind@v0.29.0

.PHONY: dev-setup
dev-setup: dev-install-kind
	KIND_EXPERIMENTAL_PROVIDER=$(KIND_EXPERIMENTAL_PROVIDER) dev-cache/kind create cluster \
		--image $(KIND_IMAGE) \
		--name $(KIND_CLUSTER_NAME) \
		--config $(KIND_CLUSTER_CONFIG_PATH) \
		--wait 5m

.PHONY: dev-teardown
dev-teardown:
	dev-cache/kind delete cluster \
		--name $(KIND_CLUSTER_NAME)

########################################################################
# Container Targets

.PHONY: image-build
image-build: image-build
	$(CONTAINER_RUNTIME) buildx build \
		-t $(REGISTRY)/$(REPOSITORY):$(TAG) \
		--platform linux/$(ARCH) -f Dockerfile .

.PHONY: image-push
image-push:
	$(info push Container image...)
	$(CONTAINER_RUNTIME) push $(REGISTRY)/$(REPOSITORY):$(TAG)

.PHONY: image-ci
image-ci: build
	$(CONTAINER_RUNTIME) buildx build \
		-t $(REGISTRY)/$(REPOSITORY):$(TAG) \
		--platform linux/$(ARCH) -f build/Containerfile-build .
