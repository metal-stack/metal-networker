package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplDroptailer is the name of the template for the droptailer service.
const tplDroptailer = "droptailer.service.tpl"

// SystemdUnitDroptailer is the name of the systemd unit for the droptailer.
const systemdUnitDroptailer = "droptailer.service"

// DroptailerData contains the data to render the droptailer service template.
type DroptailerData struct {
	Comment   string
	TenantVrf string
}

// NewDroptailerServiceApplier constructs a new instance of this type.
func NewDroptailerServiceApplier(kb KnowledgeBase, v net.Validator) (net.Applier, error) {
	tenantVrf, err := getTenantVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := DroptailerData{Comment: versionHeader(kb.Machineuuid), TenantVrf: tenantVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}

func getTenantVRFName(kb KnowledgeBase) (string, error) {
	networks := kb.GetNetworks(PrivatePrimary)
	for _, network := range networks {
		if network.Vrf != 0 {
			vrf := fmt.Sprintf("vrf%d", network.Vrf)
			return vrf, nil
		}
	}

	return "", fmt.Errorf("there is no private tenant network")
}
