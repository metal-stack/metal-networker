package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplHosts defines the name of the template to render hosts file.
const TplHosts = "hosts.tpl"

type (
	// HostsData contains data to render hosts file.
	HostsData struct {
		Comment  string
		Hostname string
		IP       string
	}

	// HostsValidator validates hosts file.
	HostsValidator struct {
		path string
	}
)

// NewHostsApplier creates a new hosts applier.
func NewHostsApplier(kb KnowledgeBase, tmpFile string) net.Applier {
	data := HostsData{Hostname: kb.Hostname, Comment: versionHeader(kb.MachineUUID), IP: kb.getPrivatePrimaryNetwork().Ips[0]}
	validator := HostsValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// Validate validates hosts file.
func (v HostsValidator) Validate() error {
	//nolint:godox
	// FIXME: How do we validate a hosts file?
	return nil
}
