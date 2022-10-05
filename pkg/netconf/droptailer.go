package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplDroptailer is the name of the template for the droptailer service.
const tplDroptailer = "droptailer.service.tpl"

// SystemdUnitDroptailer is the name of the systemd unit for the droptailer.
const systemdUnitDroptailer = "droptailer.service"

// droptailerData contains the data to render the droptailer service template.
type droptailerData struct {
	Comment   string
	TenantVrf string
}

// newDroptailerServiceApplier constructs a new instance of this type.
func newDroptailerServiceApplier(kb config, v net.Validator) (net.Applier, error) {
	tenantVrf, err := getTenantVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := droptailerData{Comment: versionHeader(kb.MachineUUID), TenantVrf: tenantVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}

func getTenantVRFName(kb config) (string, error) {
	primary := kb.getPrivatePrimaryNetwork()
	if primary.Vrf != nil && *primary.Vrf != 0 {
		vrf := fmt.Sprintf("vrf%d", *primary.Vrf)
		return vrf, nil
	}

	return "", fmt.Errorf("there is no private tenant network")
}
