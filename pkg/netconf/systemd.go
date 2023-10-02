package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// tplSystemdLinkLan defines the name of the template to render system.link file.
	tplSystemdLinkLan = "networkd/10-lan.link.tpl"

	tplSystemdNetworkLo = "networkd/00-lo.network.tpl"
	// tplSystemdNetworkLan defines the name of the template to render system.network file.
	tplSystemdNetworkLan = "networkd/10-lan.network.tpl"
	// mtuFirewall defines the value for MTU specific to the needs of a firewall. VXLAN requires higher MTU.
	mtuFirewall = 1500
	// mtuMachine defines the value for MTU specific to the needs of a machine.
	mtuMachine = 1440
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
		MAC        string
		MTU        int
		EVPNIfaces []EVPNIface
	}

	// systemdValidator validates systemd.network and system.link files.
	systemdValidator struct {
		path string
	}
)

// newSystemdNetworkdApplier creates a new Applier to configure systemd.network.
func newSystemdNetworkdApplier(tmpFile string, data any) net.Applier {
	validator := systemdValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// newSystemdLinkApplier creates a new Applier to configure systemd.link.
func newSystemdLinkApplier(kind BareMetalType, machineUUID string, nicIndex int, nic *models.V1MachineNic,
	tmpFile string, evpnIfaces []EVPNIface) (net.Applier, error) {
	var mtu int

	switch kind {
	case Firewall:
		mtu = mtuFirewall
	case Machine:
		mtu = mtuMachine
	default:
		return nil, fmt.Errorf("unknown configuratorType of configurator: %d", kind)
	}

	data := SystemdLinkData{
		SystemdCommonData: SystemdCommonData{
			Comment: versionHeader(machineUUID),
			Index:   nicIndex,
		},
		MTU:        mtu,
		MAC:        *nic.Mac,
		EVPNIfaces: evpnIfaces,
	}
	validator := systemdValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil), nil
}

// Validate validates systemd.network and systemd.link files.
func (v systemdValidator) Validate() error {
	return nil
}
