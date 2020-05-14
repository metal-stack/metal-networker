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
	go test -ldflags "-X 'github.com/metal-stack/v.Version='" -v -cover ./...

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
		-C ./bin metal-networker \
		-C ../internal/netconf/ \
			droptailer.service.tpl \
			firewall_policy_controller.service.tpl \
			frr.firewall.tpl \
			frr.machine.tpl \
			hostname.tpl \
			hosts.tpl \
			interfaces.firewall.tpl \
			lo.network.machine.tpl \
			nftables_exporter.service.tpl \
			node_exporter.service.tpl \
			rules.v4.tpl \
			rules.v6.tpl \
			suricata_config.yaml.tpl \
			suricata_defaults.tpl \
			suricata_update.service.tpl \
			systemd.link.tpl \
			systemd.network.tpl

.PHONY: validate
validate:
	./validate.sh
