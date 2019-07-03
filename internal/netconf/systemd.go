package netconf

import (
	"git.f-i-ts.de/cloud-native/metallib/network"
)

const (
	// TplSystemdLink defines the name of the template to render system.link file.
	TplSystemdLink = "systemd.link.tpl"
	// TplSystemdNetwork defines the name of the template to render system.network file.
	TplSystemdNetwork = "systemd.network.tpl"
	// MTUFirewall defines the value for MTU specific to the needs of a firewall. VXLAN requires higher MTU.
	MTUFirewall = 9216
	// MTUMachine defines the value for MTU specific to the needs of a machine.
	MTUMachine = 9000
)

// SystemdCommonData contains attributes common to systemd.network and systemd.link files.
type SystemdCommonData struct {
	Comment string
	Index   int
}

// SystemdNetworkData contains attributes required to render systemd.network files.
type SystemdNetworkData struct {
	SystemdCommonData
}

// SystemdLinkData contains attributes required to render systemd.link files.
type SystemdLinkData struct {
	SystemdCommonData
	MAC string
	MTU int
}

// SystemdValidator validates systemd.network and system.link files.
type SystemdValidator struct {
	path string
}

// NewSystemdNetworkApplier creates a new Applier to configure systemd.network.
func NewSystemdNetworkApplier(machineUUID string, nicIndex int, tmpFile string) network.Applier {
	data := SystemdNetworkData{SystemdCommonData{Comment: versionHeader(machineUUID), Index: nicIndex}}
	validator := SystemdValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
}

// NewSystemdLinkApplier creates a new Applier to configure systemd.link.
func NewSystemdLinkApplier(kind BareMetalType, machineUUID string, nicIndex int, nic NIC,
	tmpFile string) network.Applier {
	var mtu int

	switch kind {
	case Firewall:
		mtu = MTUFirewall
	case Machine:
		mtu = MTUMachine
	default:
		log.Fatalf("unknown configuratorType of configurator: %validator", kind)
	}

	data := SystemdLinkData{}
	data.Comment = versionHeader(machineUUID)
	data.MTU = mtu
	data.Index = nicIndex
	data.MAC = nic.Mac

	validator := SystemdValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
}

// Validate validates systemd.network and systemd.link files.
func (v SystemdValidator) Validate() error {
	// FIXME: We need to add a ways to validate those files.
	// https://github.com/systemd/systemd/issues/11651
	log.Infof("Skipping validation since there is no known way to validate (.network|.link) files in advance.")
	return nil
}
