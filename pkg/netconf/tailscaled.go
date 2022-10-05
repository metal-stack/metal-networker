package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// tplTailscaled is the name of the template for the tailscaled service.
	tplTailscaled = "tailscaled.service.tpl"
	// systemdUnitTailscaled is the name of the systemd unit for the tailscaled.
	systemdUnitTailscaled = "tailscaled.service"
	defaultTailscaledPort = "41641"
)

// TailscaledData contains the data to render the tailscaled service template.
type TailscaledData struct {
	TailscaledPort  string
	DefaultRouteVrf string
}

// NewTailscaledServiceApplier constructs a new instance of this type.
func NewTailscaledServiceApplier(kb config, v net.Validator) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	data := TailscaledData{TailscaledPort: defaultTailscaledPort, DefaultRouteVrf: defaultRouteVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}
