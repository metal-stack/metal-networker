.ONESHELL:
SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
BUILDDATE := $(shell date -Iseconds)
VERSION := $(or ${VERSION},devel)

# Image URL to use all building/pushing image targets
DOCKER_TAG := $(or ${GITHUB_TAG_NAME}, latest)
DOCKER_IMG ?= ghcr.io/metal-stack/metal-networker:${DOCKER_TAG}

BINARY := metal-networker

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all:: release;

.PHONY: test
test:
	GO_ENV=testing go test -v -race -cover ./...

.PHONY: all
bin/$(BINARY): test
	GGO_ENABLED=0 \
	GO111MODULE=on \
		go build \
			-trimpath \
			-tags netgo \
			-o bin/$(BINARY) \
			-ldflags "-X 'github.com/metal-stack/v.Version=$(VERSION)' \
					  -X 'github.com/metal-stack/v.Revision=$(GITVERSION)' \
					  -X 'github.com/metal-stack/v.GitSHA1=$(SHA)' \
					  -X 'github.com/metal-stack/v.BuildDate=$(BUILDDATE)'" . && strip bin/$(BINARY)

.PHONY: release
release: bin/$(BINARY) validate
	tar -czvf metal-networker.tgz \
		-C ./bin metal-networker

.PHONY: validate
validate:
	./validate.sh

# Build the docker image
docker-build:
	docker build . -t ${DOCKER_IMG}

# Push the docker image
docker-push:
	docker push ${DOCKER_IMG}
