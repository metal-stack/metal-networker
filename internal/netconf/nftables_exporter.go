package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplNftablesExporter is the name of the template for the nftables_exporter service.
const TplNftablesExporter = "nftables_exporter.service.tpl"

// SystemdUnitNftablesExporter is the name of the systemd unit for the nftables_exporter.
const SystemdUnitNftablesExporter = "nftables-exporter.service"

// NftablesExporterData contains the data to render the nftables_exporter service template.
type NftablesExporterData struct {
	Comment   string
	TenantVrf string
}

// NewNftablesExporterServiceApplier constructs a new instance of this type.
func NewNftablesExporterServiceApplier(kb KnowledgeBase, v net.Validator) (net.Applier, error) {
	tenantVrf, err := getTenantVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := NftablesExporterData{Comment: versionHeader(kb.Machineuuid), TenantVrf: tenantVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}
