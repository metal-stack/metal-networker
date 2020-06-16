package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplFirewallController is the name of the template for the firewall-policy-controller service.
const TplFirewallController = "firewall_controller.service.tpl"

// SystemdUnitFirewallController is the name of the systemd unit for the firewall policy controller,
const SystemdUnitFirewallController = "firewall-controller.service"

// FirewallControllerData contains the data to render the firewall-controller service template.
type FirewallControllerData struct {
	Comment         string
	DefaultRouteVrf string
}

// NewFirewallControllerServiceApplier constructs a new instance of this type.
func NewFirewallControllerServiceApplier(kb KnowledgeBase, v net.Validator) (net.Applier, error) {
	defaultRouteVrf, err := getDefaultRouteVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := FirewallControllerData{Comment: versionHeader(kb.Machineuuid), DefaultRouteVrf: defaultRouteVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}

// ServiceValidator holds information for systemd service validation.
type ServiceValidator struct {
	path string
}

// Validate validates the service file.
func (v ServiceValidator) Validate() error {
	// Currently not implemented as systemd-analyze fails in the metal-hammer.
	// Error: Cannot determine cgroup we are running in: No medium found
	return nil
}