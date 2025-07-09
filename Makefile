# This project applies to ppc64le only
ARCH ?= ppc64le

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

########################################################################
# Go Targets

build: fmt vet
	GOOS=linux GOARCH=$(ARCH) go build -o bin/power-dra-kubeletplugin cmd/power-dra-kubeletplugin/*.go

controller-gen: ## Download controller-gen locally if necessary.
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0
CONTROLLER_GEN=$(shell go env GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

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

binaries: cmds
ifneq ($(PREFIX),)
cmd-%: COMMAND_BUILD_OPTIONS = -o $(PREFIX)/$(*)
endif
cmds: $(CMD_TARGETS)
$(CMD_TARGETS): cmd-%:
	CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' \
		CC=$(CC) CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "-s -w -X $(CLI_VERSION_PACKAGE).gitCommit=$(GIT_COMMIT) -X $(CLI_VERSION_PACKAGE).version=$(VERSION)" $(COMMAND_BUILD_OPTIONS) $(MODULE)/cmd/$(*)

.PHONY: dev-setup
dev-setup:
	dev/setup.sh

.PHONY: dev-clean
dev-clean:
	dev/teardown.sh

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
