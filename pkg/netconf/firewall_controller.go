package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplFirewallController is the name of the template for the firewall-policy-controller service.
const tplFirewallController = "firewall_controller.service.tpl"

// SystemdUnitFirewallController is the name of the systemd unit for the firewall policy controller,
const systemdUnitFirewallController = "firewall-controller.service"

// firewallControllerData contains the data to render the firewall-controller service template.
type firewallControllerData struct {
	Comment         string
	DefaultRouteVrf string
	ServiceIP       string
	PrivateVrfID    int64
}

// newFirewallControllerServiceApplier constructs a new instance of this type.
func newFirewallControllerServiceApplier(kb config, v net.Validator) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	if len(kb.getPrivatePrimaryNetwork().Ips) == 0 {
		return nil, fmt.Errorf("no private IP found useable for the firewall controller")
	}
	data := firewallControllerData{
		Comment:         versionHeader(kb.MachineUUID),
		DefaultRouteVrf: defaultRouteVrf,
	}

	return net.NewNetworkApplier(data, v, nil), nil
}

// serviceValidator holds information for systemd service validation.
type serviceValidator struct {
	path string
}

// Validate validates the service file.
func (v serviceValidator) Validate() error {
	// Currently not implemented as systemd-analyze fails in the metal-hammer.
	// Error: Cannot determine cgroup we are running in: No medium found
	return nil
}
