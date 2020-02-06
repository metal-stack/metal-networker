.ONESHELL:
SHA := $(shell git rev-parse --short=8 HEAD)
GITVERSION := $(shell git describe --long --all)
BUILDDATE := $(shell date -Iseconds)
VERSION := $(or ${VERSION},devel)

BINARY := metal-networker

.PHONY: all
all:: release;

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: all
bin/$(BINARY): test
	GGO_ENABLED=0 \
	GO111MODULE=on \
		go build \
			-trimpath \
			-tags netgo \
			-o bin/$(BINARY) \
			-ldflags "-X 'github.com/metal-pod/v.Version=$(VERSION)' \
					  -X 'github.com/metal-pod/v.Revision=$(GITVERSION)' \
					  -X 'github.com/metal-pod/v.GitSHA1=$(SHA)' \
					  -X 'github.com/metal-pod/v.BuildDate=$(BUILDDATE)'" . && strip bin/$(BINARY)

.PHONY: release
release: bin/$(BINARY) validate
	tar -czvf metal-networker.tgz \
		-C ./bin metal-networker \
		-C ../internal/netconf/ \
			interfaces.firewall.tpl \
			interfaces.machine.tpl \
			frr.machine.tpl \
			frr.firewall.tpl \
			rules.v4.tpl \
			rules.v6.tpl \
			systemd.link.tpl \
			systemd.network.tpl \
			hosts.tpl \
			hostname.tpl \
			droptailer.service.tpl \
			firewall_policy_controller.service.tpl \
			nftables_exporter.service.tpl \
			node_exporter.service.tpl

.PHONY: validate
validate:
	./validate.sh