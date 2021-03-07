package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// tplEveboxAgent is the name of the template for the evebox-agent service.
const tplEveboxAgent = "evebox-agent.service.tpl"

// systemdUnitEveboxAgent is the name of the systemd unit for the evebox-agent.
const systemdUnitEveboxAgent = "evebox-agent.service.service"

// EveboxAgentData contains the data to render the evebox-agent service template.
type EveboxAgentData struct {
	Comment         string
	DefaultRouteVrf string
}

// NewEveboxAgentServiceApplier constructs a new instance of this type.
func NewEveboxAgentServiceApplier(kb KnowledgeBase, v net.Validator) (net.Applier, error) {
	defaultRouteVrf, err := getDefaultRouteVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := EveboxAgentData{Comment: versionHeader(kb.Machineuuid), DefaultRouteVrf: defaultRouteVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}
