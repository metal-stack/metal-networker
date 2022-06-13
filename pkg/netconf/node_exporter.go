package netconf

import (
	"github.com/metal-stack/metal-networker/pkg/net"
)

// tplNodeExporter is the name of the template for the node_exporter service.
const tplNodeExporter = "node_exporter.service.tpl"

// systemdUnitNodeExporter is the name of the systemd unit for the node_exporter.
const systemdUnitNodeExporter = "node-exporter.service"

// NodeExporterData contains the data to render the node_exporter service template.
type NodeExporterData struct {
	Comment   string
	TenantVrf string
}

// NewNodeExporterServiceApplier constructs a new instance of this type.
func NewNodeExporterServiceApplier(kb KnowledgeBase, v net.Validator) (*net.NetworkApplier, error) {
	tenantVrf, err := getTenantVRFName(kb)
	if err != nil {
		return nil, err
	}

	data := NodeExporterData{Comment: versionHeader(kb.Machineuuid), TenantVrf: tenantVrf}

	return net.NewNetworkApplier(data, v, nil), nil
}
