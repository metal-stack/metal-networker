package netconf

import (
	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"
	"git.f-i-ts.de/cloud-native/metallib/network"
)

// TplFirewallPolicyControllerService is the name of the template for the firewall-policy-controller service.
const TplFirewallPolicyControllerService = "firewall_policy_controller.service.tpl"

// FirewallPolicyControllerData contains the data to render the firewall-policy-controller service template.
type FirewallPolicyControllerData struct {
	Comment         string
	DefaultRouteVrf string
}

// NewFirewallPolicyControllerServiceApplier constructs a new instance of this type.
func NewFirewallPolicyControllerServiceApplier(kb KnowledgeBase, v ServiceValidator) (network.Applier, error) {
	defaultRouteVrf, err := getDefaultRouteVRFName(kb)
	if err != nil {
		return nil, err
	}
	data := FirewallPolicyControllerData{Comment: versionHeader(kb.Machineuuid), DefaultRouteVrf: defaultRouteVrf}
	return network.NewNetworkApplier(data, v, nil), nil
}

// ServiceValidator holds information for systemd service validation.
type ServiceValidator struct {
	path string
}

// Validate validates the service file.
func (v ServiceValidator) Validate() error {
	log.Infof("running 'systemd-analyze verify %s to validate changes.'", v.path)
	return exec.NewVerboseCmd("systemd-analyze", "verify", v.path).Run()
}
