package netconf

import (
	"strings"

	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplSuricataConfig is the name of the template for the suricata configuration.
const TplSuricataConfig = "suricata_config.yaml.tpl"

// SuricataConfigData represents the information required to render suricata configuration.
type SuricataConfigData struct {
	Comment         string
	DefaultRouteVrf string
	Interface       string
}

// SuricataConfigValidator can validate configuration for suricata.
type SuricataConfigValidator struct {
	path string
}

// NewSuricataConfigApplier constructs a new instance of this type.
func NewSuricataConfigApplier(kb KnowledgeBase, tmpFile string) (net.Applier, error) {
	defaultRouteVrf, err := kb.getDefaultRouteVRFName()
	if err != nil {
		return nil, err
	}

	i := strings.Replace(defaultRouteVrf, "vrf", "vlan", 1)
	data := SuricataConfigData{Comment: versionHeader(kb.MachineUUID), DefaultRouteVrf: defaultRouteVrf, Interface: i}
	validator := SuricataConfigValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil), nil
}

// Validate validates suricata configuration.
func (v SuricataConfigValidator) Validate() error {
	return nil
}
