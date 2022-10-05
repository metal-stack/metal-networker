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

// suricataDefaultsValidator can validate defaults for suricata.
type suricataDefaultsValidator struct {
	path string
}

// newSuricataDefaultsApplier constructs a new instance of this type.
func newSuricataDefaultsApplier(kb config, tmpFile string) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	i := strings.Replace(defaultRouteVrf, "vrf", "vlan", 1)
	data := SuricataDefaultsData{Comment: versionHeader(kb.MachineUUID), Interface: i}
	validator := suricataDefaultsValidator{path: tmpFile}

	return net.NewNetworkApplier(data, validator, nil), nil
}

// Validate validates suricata defaults.
func (v suricataDefaultsValidator) Validate() error {
	return nil
}
