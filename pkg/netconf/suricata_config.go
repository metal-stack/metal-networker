package netconf

import (
	"strings"

	"github.com/metal-stack/metal-networker/pkg/net"
)

// tplSuricataConfig is the name of the template for the suricata configuration.
const tplSuricataConfig = "suricata_config.yaml.tpl"

// SuricataConfigData represents the information required to render suricata configuration.
type SuricataConfigData struct {
	Comment         string
	DefaultRouteVrf string
	Interface       string
}

// suricataConfigValidator can validate configuration for suricata.
type suricataConfigValidator struct {
	path string
}

// newSuricataConfigApplier constructs a new instance of this type.
func newSuricataConfigApplier(kb config, tmpFile string) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	i := strings.Replace(defaultRouteVrf, "vrf", "vlan", 1)
	data := SuricataConfigData{Comment: versionHeader(kb.MachineUUID), DefaultRouteVrf: defaultRouteVrf, Interface: i}
	validator := suricataConfigValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil), nil
}

// Validate validates suricata configuration.
func (v suricataConfigValidator) Validate() error {
	return nil
}
