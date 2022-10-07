.ONESHELL:
SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
BUILDDATE := $(shell date -Iseconds)
VERSION := $(or ${VERSION},devel)

.PHONY: all
all: test validate

.PHONY: test
test:
	GO_ENV=testing go test -v -race -cover ./...

.PHONY: validate
validate:
	./validate.sh
