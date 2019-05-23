.ONESHELL:
BINARY := metal-networker
COMMONDIR := $(or ${COMMONDIR},../common)
SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
BUILDDATE := $(shell date -Iseconds)
VERSION := $(or ${VERSION},devel)

include $(COMMONDIR)/Makefile.inc

${BINARY}: clean test
    GGO_ENABLED=0 \
    GO111MODULE=on \
    go build \
    -tags netgo \
    -ldflags "-X 'git.f-i-ts.de/cloud-native/metallib/version.Version=$(VERSION)' \
              -X 'git.f-i-ts.de/cloud-native/metallib/version.Revision=$(GITVERSION)' \
              -X 'git.f-i-ts.de/cloud-native/metallib/version.Gitsha1=$(SHA)' \
              -X 'git.f-i-ts.de/cloud-native/metallib/version.Builddate=$(BUILDDATE)'" \
    -o metal-networker
