package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// tplSuricataUpdate is the name of the template for the suricata-update service.
const tplSuricataUpdate = "suricata_update.service.tpl"

// systemdUnitSuricataUpdate is the name of the systemd unit for the suricata-update.
const systemdUnitSuricataUpdate = "suricata-update.service"

// SuricataUpdateData contains the data to render the suricata-update service template.
type SuricataUpdateData struct {
	Comment         string
	DefaultRouteVrf string
}

// NewSuricataUpdateServiceApplier constructs a new instance of this type.
func NewSuricataUpdateServiceApplier(kb KnowledgeBase, v net.Validator) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	data := SuricataUpdateData{Comment: versionHeader(kb.MachineUUID), DefaultRouteVrf: defaultRouteVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}
