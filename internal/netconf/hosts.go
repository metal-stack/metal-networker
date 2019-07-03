package netconf

import "git.f-i-ts.de/cloud-native/metallib/network"

// TplHosts defines the name of the template to render hosts file.
const TplHosts = "hosts.tpl"

// HostsData contains data to render hosts file.
type HostsData struct {
	Comment  string
	Hostname string
	IP       string
}

// HostsValidator validates hosts file.
type HostsValidator struct {
	path string
}

// NewHostsApplier creates a new hosts applier.
func NewHostsApplier(kb KnowledgeBase, tmpFile string) network.Applier {
	data := HostsData{Hostname: kb.Hostname, Comment: versionHeader(kb.Machineuuid), IP: kb.getPrimaryNetwork().Ips[0]}
	validator := HostsValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
}

// Validate validates hosts file.
func (v HostsValidator) Validate() error {
	// FIXME: How do we validate a hosts file?
	return nil
}
