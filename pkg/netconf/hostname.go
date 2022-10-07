package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// tplHostname defines the name of the template to render /etc/hostname.
const tplHostname = "hostname.tpl"

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

// newHostnameApplier creates a new Applier to render hostname.
func newHostnameApplier(kb config, tmpFile string) net.Applier {
	data := HostnameData{Comment: versionHeader(kb.MachineUUID), Hostname: kb.Hostname}
	validator := HostnameValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// Validate validates hostname rendering.
func (v HostnameValidator) Validate() error {
	return nil
}
