package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// tplSuricata is the name of the template for the suricata service.
const tplSuricata = "suricata.service.tpl"

// systemdUnitSuricata is the name of the systemd unit for the suricata.
const systemdUnitSuricata = "suricata.service"

// SuricataData contains the data to render the suricata service template.
type SuricataData struct {
	EnableIPS bool
}

// NewSuricataServiceApplier constructs a new instance of this type.
func NewSuricataServiceApplier(kb KnowledgeBase, v net.Validator, enableIPS bool) (net.Applier, error) {
	data := SuricataData{EnableIPS: enableIPS}
	return net.NewNetworkApplier(data, v, nil), nil
}

func GetSystemdUnitConfig(enableIPS bool) UnitConfiguration {
	return UnitConfiguration{
		unit:         systemdUnitSuricata,
		templateFile: tplSuricata,
		constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
			return NewSuricataServiceApplier(kb, v, enableIPS)
		},
	}
}
