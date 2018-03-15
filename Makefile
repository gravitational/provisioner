# Utility script to build, test and push image to registry
CWD := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

VENDORBIN = $(CURDIR)/vendor/bin
PATH     := $(VENDORBIN):$(PATH)

PROV_VERSION ?= $(shell git describe --tags 2>/dev/null ||  git rev-parse HEAD)
PROV_REPO = quay.io/gravitational/provisioner
TERRAFORM_VER ?= 0.11.4
BUILDBOX_TAG ?= golang:1.9.0-stretch

# Build docker image
.PHONY: build-provisioner
build-provisioner: build
	docker build \
           --build-arg=TERRAFORM_VER=$(TERRAFORM_VER) \
           -t "$(PROV_REPO):$(PROV_VERSION)" .

# Publish docker image. User runs this has to have Quay write permission
.PHONY: publish-provisioner
publish-provisioner:
	docker push $(PROV_REPO):$(PROV_VERSION)

.PHONY: deps
deps: vendor

vendor: Gopkg.lock
	dep ensure

# builds privisioner program inside Docker container
.PHONY: build
build: deps
	mkdir -p $(CWD)/build
	docker run -v $(CWD)/build:/build -v $(CWD):/go/src/github.com/gravitational/provisioner $(BUILDBOX_TAG) go build -o /build/provisioner github.com/gravitational/provisioner/cmd

# Run go test in docker
.PHONY: test
test: deps
	docker run -v $(CWD):/go/src/github.com/gravitational/provisioner $(BUILDBOX_TAG) go test github.com/gravitational/provisioner/...
