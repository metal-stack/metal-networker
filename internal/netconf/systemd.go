package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
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

type (
	// SystemdCommonData contains attributes common to systemd.network and systemd.link files.
	SystemdCommonData struct {
		Comment string
		Index   int
	}

	// SystemdNetworkData contains attributes required to render systemd.network files.
	SystemdNetworkData struct {
		SystemdCommonData
	}

	// SystemdLinkData contains attributes required to render systemd.link files.
	SystemdLinkData struct {
		SystemdCommonData
		MAC string
		MTU int
	}

	// SystemdValidator validates systemd.network and system.link files.
	SystemdValidator struct {
		path string
		log  *zap.SugaredLogger
	}
)

// NewSystemdNetworkApplier creates a new Applier to configure systemd.network.
func NewSystemdNetworkApplier(uuid string, nicIndex int, tmpFile string, log *zap.SugaredLogger) net.Applier {
	data := SystemdNetworkData{SystemdCommonData{Comment: versionHeader(uuid), Index: nicIndex}}
	validator := SystemdValidator{tmpFile, log}

	return net.NewNetworkApplier(data, validator, nil)
}

// NewSystemdLinkApplier creates a new Applier to configure systemd.link.
func NewSystemdLinkApplier(kind BareMetalType, machineUUID string, nicIndex int, nic NIC,
	tmpFile string, log *zap.SugaredLogger) net.Applier {
	var mtu int

	switch kind {
	case Firewall:
		mtu = MTUFirewall
	case Machine:
		mtu = MTUMachine
	default:
		log.Fatalf("unknown configuratorType of configurator: %validator", kind)
	}

	data := SystemdLinkData{
		SystemdCommonData: SystemdCommonData{
			Comment: versionHeader(machineUUID),
			Index:   nicIndex,
		},
		MTU: mtu,
		MAC: nic.Mac,
	}
	validator := SystemdValidator{tmpFile, log}

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
