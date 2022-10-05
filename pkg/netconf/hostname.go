package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplHostname defines the name of the template to render /etc/hostname.
const TplHostname = "hostname.tpl"

type (
	// HostnameData contains attributes to render hostname file.
	HostnameData struct {
		Comment, Hostname string
	}

	// HostnameValidator validates hostname changes.
	HostnameValidator struct {
		path string
	}
)

// NewHostnameApplier creates a new Applier to render hostname.
func NewHostnameApplier(kb KnowledgeBase, tmpFile string) net.Applier {
	data := HostnameData{Comment: versionHeader(kb.MachineUUID), Hostname: kb.Hostname}
	validator := HostnameValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// Validate validates hostname rendering.
func (v HostnameValidator) Validate() error {
	return nil
}
