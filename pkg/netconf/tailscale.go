package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// tplTailscale is the name of the template for the Tailscale client.
	tplTailscale = "tailscale.service.tpl"
	// systemdUnitTailscale is the name of the systemd unit for the Tailscale client.
	systemdUnitTailscale = "tailscale.service"
)

// TailscaleData contains the data to render the Tailscale service template.
type TailscaleData struct {
	MachineID       string
	AuthKey         string
	Address         string
	DefaultRouteVrf string
}

// NewTailscaleServiceApplier constructs a new instance of this type.
func NewTailscaleServiceApplier(kb config, v net.Validator) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	data := TailscaleData{
		MachineID:       kb.MachineUUID,
		AuthKey:         *kb.VPN.AuthKey,
		Address:         *kb.VPN.Address,
		DefaultRouteVrf: defaultRouteVrf,
	}

	return net.NewNetworkApplier(data, v, nil), nil
}
