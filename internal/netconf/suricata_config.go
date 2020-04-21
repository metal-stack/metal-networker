package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplSuricataConfig is the name of the template for the suricata configuration.
const TplSuricataConfig = "suricata_config.yaml.tpl"

// SuricataConfigData represents the information required to render suricata configuration.
type SuricataConfigData struct {
	Comment         string
	DefaultRouteVrf string
}

// SuricataConfigValidator can validate configuration for suricata.
type SuricataConfigValidator struct {
	path string
}

// NewSuricataConfigApplier constructs a new instance of this type.
func NewSuricataConfigApplier(kb KnowledgeBase, tmpFile string) (net.Applier, error) {
	defaultRouteVrf, err := getDefaultRouteVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := SuricataUpdateData{Comment: versionHeader(kb.Machineuuid), DefaultRouteVrf: defaultRouteVrf}
	validator := SuricataConfigValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil), nil
}

// Validate validates suricata configuration.
func (v SuricataConfigValidator) Validate() error {
	return nil
}
