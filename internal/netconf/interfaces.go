package netconf

import (
	"fmt"
	"io"
	"text/template"
)

type (
	// IfacesData contains attributes required to render network interfaces configuration of a bare metal
	// server.
	IfacesData struct {
		Comment  string
		Loopback Loopback
		Tenants  []Tenant
	}

	// SystemdNetworkdValidator defines the base type of an systemd-networkd validator.
	SystemdNetworkdValidator struct {
		path string
	}
)

type IfacesApplier struct {
	kind BareMetalType
	kb   KnowledgeBase
	d    IfacesData
}

// NewIfacesConfigApplier constructs a new instance of this type.
func NewIfacesConfigApplier(kind BareMetalType, kb KnowledgeBase) IfacesApplier {
	d := IfacesData{
		Comment: versionHeader(kb.Machineuuid),
	}

	switch kind {
	case Firewall:
		underlay := kb.getUnderlayNetwork()
		d.Loopback.Comment = fmt.Sprintf("# networkid: %s", underlay.Networkid)
		d.Loopback.IPs = underlay.Ips
		d.Tenants = getTenants(kb)
	case Machine:
		private := kb.getPrivateNetwork()
		d.Loopback.Comment = fmt.Sprintf("# networkid: %s", private.Networkid)
		// Ensure that the ips of the private network are the first ips at the loopback interface.
		// The first lo IP is used within network communication and other systems depend on seeing the first private ip.
		d.Loopback.IPs = append(private.Ips, kb.CollectIPs(Public)...)
	default:
		log.Fatalf("unknown configuratorType of configurator: %v", kind)
	}

	return IfacesApplier{kind: kind, kb: kb, d: d}
}

// Render renders the network interfaces to the given writer using the given template.
func (a *IfacesApplier) Render(w io.Writer, tpl template.Template) error {
	return tpl.Execute(w, a.d)
}

// Apply applies the interface configuration with systemd-networkd.
func (a *IfacesApplier) Apply() {
	uuid := a.kb.Machineuuid
	tenants := a.d.Tenants

	// /etc/systemd/network/00 loopback
	src := mustTmpFile("lo_network_")
	applier := NewSystemdNetworkdApplier(src, a.d)
	dest := fmt.Sprintf("%s/00-lo.network", SystemdNetworkPath)
	applyAndCleanUp(applier, tplSystemdNetworkLo, src, dest, FileModeSystemd)

	// /etc/systemd/network/1x* lan interfaces
	offset := 10
	for i, nic := range a.kb.Nics {
		prefix := fmt.Sprintf("lan%d_link_", i)
		src := mustTmpFile(prefix)
		applier := NewSystemdLinkApplier(a.kind, uuid, i, nic, src, tenants)
		dest := fmt.Sprintf("%s/%d-lan%d.link", SystemdNetworkPath, offset+i, i)
		applyAndCleanUp(applier, tplSystemdLinkLan, src, dest, FileModeSystemd)

		prefix = fmt.Sprintf("lan%d_network_", i)
		src = mustTmpFile(prefix)
		applier = NewSystemdLinkApplier(a.kind, uuid, i, nic, src, tenants)
		dest = fmt.Sprintf("%s/%d-lan%d.network", SystemdNetworkPath, offset+i, i)
		applyAndCleanUp(applier, tplSystemdNetworkLan, src, dest, FileModeSystemd)
	}

	if a.kind == Machine {
		return
	}

	// /etc/systemd/network/20 bridge interface
	applyNetdevAndNetwork(20, 20, "bridge", "", a.d)

	// /etc/systemd/network/3x* triplet of interfaces for a tenant: vrf, svi, vxlan
	offset = 30
	for i, tenant := range a.d.Tenants {
		suffix := fmt.Sprintf("-%d", tenant.VRF.ID)
		applyNetdevAndNetwork(offset, offset+i, "vrf", suffix, tenant)
		applyNetdevAndNetwork(offset, offset+i, "svi", suffix, tenant)
		applyNetdevAndNetwork(offset, offset+i, "vxlan", suffix, tenant)
	}
}

func applyNetdevAndNetwork(si, di int, prefix, suffix string, data interface{}) {
	src := mustTmpFile(prefix + "_netdev_")
	applier := NewSystemdNetworkdApplier(src, data)
	dest := fmt.Sprintf("%s/%d-%s%s.netdev", SystemdNetworkPath, di, prefix, suffix)
	tpl := fmt.Sprintf("networkd/%d-%s.netdev.tpl", si, prefix)
	applyAndCleanUp(applier, tpl, src, dest, FileModeSystemd)

	if prefix == "vrf" {
		return
	}
	src = mustTmpFile(prefix + "_network_")
	applier = NewSystemdNetworkdApplier(src, data)
	dest = fmt.Sprintf("%s/%d-%s%s.network", SystemdNetworkPath, di, prefix, suffix)
	tpl = fmt.Sprintf("networkd/%d-%s.network.tpl", si, prefix)
	applyAndCleanUp(applier, tpl, src, dest, FileModeSystemd)
}

// Validate network interfaces configuration done with systemd-networkd. Assumes systemd-networkd is installed.
func (v SystemdNetworkdValidator) Validate() error {
	log.Infof("systemd-networkd does not have validation capabilities for the .network file: %s", v.path)
	return nil
}

func getTenants(kb KnowledgeBase) []Tenant {
	var result []Tenant

	offset := 1000
	for i, n := range kb.Networks {
		if n.Underlay {
			continue
		}

		e := Tenant{}
		e.Comment = versionHeader(kb.Machineuuid)
		e.SVI.Comment = fmt.Sprintf("# svi (networkid: %s)", n.Networkid)
		e.SVI.VLANID = n.Vlan
		e.SVI.Addresses = n.Ips
		e.VXLAN.Comment = fmt.Sprintf("# vxlan (networkid: %s)", n.Networkid)
		e.VXLAN.ID = n.Vrf
		e.VXLAN.TunnelIP = kb.getUnderlayNetwork().Ips[0]
		e.VRF.Comment = fmt.Sprintf("# vrf (networkid: %s)", n.Networkid)
		e.VRF.ID = n.Vrf
		e.VRF.Table = offset + i
		result = append(result, e)
	}

	return result
}
