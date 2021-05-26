package netconf

import (
	"strings"

	"github.com/metal-stack/metal-networker/pkg/net"
)

// tplSuricataDefaults is the name of the template for the suricata defaults.
const tplSuricataDefaults = "suricata_defaults.tpl"

// SuricataDefaultsData represents the information required to render suricata defaults.
type SuricataDefaultsData struct {
	Comment   string
	Interface string
}

// SuricataDefaultsValidator can validate defaults for suricata.
type SuricataDefaultsValidator struct {
	path string
}

// NewSuricataDefaultsApplier constructs a new instance of this type.
func NewSuricataDefaultsApplier(kb KnowledgeBase, tmpFile string) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	i := strings.Replace(defaultRouteVrf, "vrf", "vlan", 1)
	data := SuricataDefaultsData{Comment: versionHeader(kb.Machineuuid), Interface: i}
	validator := SuricataDefaultsValidator{path: tmpFile}

	return net.NewNetworkApplier(data, validator, nil), nil
}

// Validate validates suricata defaults.
func (v SuricataDefaultsValidator) Validate() error {
	return nil
}
