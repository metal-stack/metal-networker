package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplSuricataUpdate is the name of the template for the suricata-update service.
const TplSuricataUpdate = "suricata_update.service.tpl"

// SystemdUnitSuricataUpdate is the name of the systemd unit for the suricata-update.
const SystemdUnitSuricataUpdate = "suricata-update.service"

// SuricataUpdateData contains the data to render the suricata-update service template.
type SuricataUpdateData struct {
	Comment         string
	DefaultRouteVrf string
}

// NewSuricataUpdateServiceApplier constructs a new instance of this type.
func NewSuricataUpdateServiceApplier(kb KnowledgeBase, v net.Validator) (net.Applier, error) {
	defaultRouteVrf, err := getDefaultRouteVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := SuricataUpdateData{Comment: versionHeader(kb.Machineuuid), DefaultRouteVrf: defaultRouteVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}
