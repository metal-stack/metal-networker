package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// tplSystemdLinkLan defines the name of the template to render system.link file.
	tplSystemdLinkLan = "networkd/10-lan.link.tpl"

	tplSystemdNetworkLo = "networkd/00-lo.network.tpl"
	// tplSystemdNetworkLan defines the name of the template to render system.network file.
	tplSystemdNetworkLan = "networkd/10-lan.network.tpl"
	// mtuFirewall defines the value for MTU specific to the needs of a firewall. VXLAN requires higher MTU.
	mtuFirewall = 9216
	// mtuMachine defines the value for MTU specific to the needs of a machine.
	mtuMachine = 9000
)

type (
	// SystemdCommonData contains attributes common to systemd.network and systemd.link files.
	SystemdCommonData struct {
		Comment string
		Index   int
	}

	// SystemdLinkData contains attributes required to render systemd.link files.
	SystemdLinkData struct {
		SystemdCommonData
		MAC     string
		MTU     int
		Tenants []Tenant
	}

	// SystemdValidator validates systemd.network and system.link files.
	SystemdValidator struct {
		path string
	}
)

// NewSystemdNetworkdApplier creates a new Applier to configure systemd.network.
func NewSystemdNetworkdApplier(tmpFile string, data interface{}) net.Applier {
	validator := SystemdValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// NewSystemdLinkApplier creates a new Applier to configure systemd.link.
func NewSystemdLinkApplier(kind BareMetalType, machineUUID string, nicIndex int, nic NIC,
	tmpFile string, tenants []Tenant) net.Applier {
	var mtu int

	switch kind {
	case Firewall:
		mtu = mtuFirewall
	case Machine:
		mtu = mtuMachine
	default:
		log.Fatalf("unknown configuratorType of configurator: %validator", kind)
	}

	data := SystemdLinkData{
		SystemdCommonData: SystemdCommonData{
			Comment: versionHeader(machineUUID),
			Index:   nicIndex,
		},
		MTU:     mtu,
		MAC:     nic.Mac,
		Tenants: tenants,
	}
	validator := SystemdValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// Validate validates systemd.network and systemd.link files.
func (v SystemdValidator) Validate() error {
	//nolint:godox
	// FIXME: We need to add a way to validate those files.
	// https://github.com/systemd/systemd/issues/11651
	log.Infof("Skipping validation since there is no known way to validate (.network|.link) files in advance.")
	return nil
}
