package netconf

import "git.f-i-ts.de/cloud-native/metallib/network"

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
func NewHostnameApplier(kb KnowledgeBase, tmpFile string) network.Applier {
	data := HostnameData{Comment: versionHeader(kb.Machineuuid), Hostname: kb.Hostname}
	validator := HostnameValidator{tmpFile}

	return network.NewNetworkApplier(data, validator, nil)
}

// Validate validates hostname rendering.
func (v HostnameValidator) Validate() error {
	return nil
}
