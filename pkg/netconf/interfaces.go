package netconf

import (
	"fmt"
	"io"
	"net/netip"
	"text/template"

	mn "github.com/metal-stack/metal-lib/pkg/net"
	"go.uber.org/zap"
)

type (
	// IfacesData contains attributes required to render network interfaces configuration of a bare metal
	// server.
	IfacesData struct {
		Comment    string
		Loopback   Loopback
		EVPNIfaces []EVPNIface
	}
)

// ifacesApplier applies interfaces configuration.
type ifacesApplier struct {
	kind BareMetalType
	kb   config
	data IfacesData
}

// newIfacesApplier constructs a new instance of this type.
func newIfacesApplier(kind BareMetalType, c config) ifacesApplier {
	d := IfacesData{
		Comment: versionHeader(c.MachineUUID),
	}

	switch kind {
	case Firewall:
		underlay := c.getUnderlayNetwork()
		d.Loopback.Comment = fmt.Sprintf("# networkid: %s", *underlay.Networkid)
		d.Loopback.IPs = addBitlen(underlay.Ips)
		d.EVPNIfaces = getEVPNIfaces(c)
	case Machine:
		private := c.getPrivatePrimaryNetwork()
		d.Loopback.Comment = fmt.Sprintf("# networkid: %s", *private.Networkid)
		// Ensure that the ips of the private network are the first ips at the loopback interface.
		// The first lo IP is used within network communication and other systems depend on seeing the first private ip.
		d.Loopback.IPs = addBitlen(append(private.Ips, c.CollectIPs(mn.External)...))
	default:
		c.log.Fatalf("unknown configuratorType of configurator: %v", kind)
	}

	return ifacesApplier{kind: kind, kb: c, data: d}
}

func addBitlen(ips []string) []string {
	ipsWithMask := []string{}
	for _, ip := range ips {
		parsedIP, err := netip.ParseAddr(ip)
		if err != nil {
			continue
		}
		ipWithMask := fmt.Sprintf("%s/%d", ip, parsedIP.BitLen())
		ipsWithMask = append(ipsWithMask, ipWithMask)
	}
	return ipsWithMask
}

// Render renders the network interfaces to the given writer using the given template.
func (a *ifacesApplier) Render(w io.Writer, tpl template.Template) error {
	return tpl.Execute(w, a.data)
}

// Apply applies the interface configuration with systemd-networkd.
func (a *ifacesApplier) Apply() {
	uuid := a.kb.MachineUUID
	evpnIfaces := a.data.EVPNIfaces

	// /etc/systemd/network/00 loopback
	src := mustTmpFile("lo_network_")
	applier := newSystemdNetworkdApplier(src, a.data)
	dest := fmt.Sprintf("%s/00-lo.network", systemdNetworkPath)
	applyAndCleanUp(a.kb.log, applier, tplSystemdNetworkLo, src, dest, fileModeSystemd, false)

	// /etc/systemd/network/1x* lan interfaces
	offset := 10
	for i, nic := range a.kb.Nics {
		prefix := fmt.Sprintf("lan%d_link_", i)
		src := mustTmpFile(prefix)
		applier, err := newSystemdLinkApplier(a.kind, uuid, i, nic, src, evpnIfaces)
		if err != nil {
			a.kb.log.Fatalw("unable to create systemdlinkapplier", "error", err)
		}
		dest := fmt.Sprintf("%s/%d-lan%d.link", systemdNetworkPath, offset+i, i)
		applyAndCleanUp(a.kb.log, applier, tplSystemdLinkLan, src, dest, fileModeSystemd, false)

		prefix = fmt.Sprintf("lan%d_network_", i)
		src = mustTmpFile(prefix)
		applier, err = newSystemdLinkApplier(a.kind, uuid, i, nic, src, evpnIfaces)
		if err != nil {
			a.kb.log.Fatalw("unable to create systemdlinkapplier", "error", err)
		}
		dest = fmt.Sprintf("%s/%d-lan%d.network", systemdNetworkPath, offset+i, i)
		applyAndCleanUp(a.kb.log, applier, tplSystemdNetworkLan, src, dest, fileModeSystemd, false)
	}

	if a.kind == Machine {
		return
	}

	// /etc/systemd/network/20 bridge interface
	applyNetdevAndNetwork(a.kb.log, 20, 20, "bridge", "", a.data)

	// /etc/systemd/network/3x* triplet of interfaces for a tenant: vrf, svi, vxlan
	offset = 30
	for i, tenant := range a.data.EVPNIfaces {
		suffix := fmt.Sprintf("-%d", tenant.VRF.ID)
		applyNetdevAndNetwork(a.kb.log, offset, offset+i, "vrf", suffix, tenant)
		applyNetdevAndNetwork(a.kb.log, offset, offset+i, "svi", suffix, tenant)
		applyNetdevAndNetwork(a.kb.log, offset, offset+i, "vxlan", suffix, tenant)
	}
}

func applyNetdevAndNetwork(log *zap.SugaredLogger, si, di int, prefix, suffix string, data any) {
	src := mustTmpFile(prefix + "_netdev_")
	applier := newSystemdNetworkdApplier(src, data)
	dest := fmt.Sprintf("%s/%d-%s%s.netdev", systemdNetworkPath, di, prefix, suffix)
	tpl := fmt.Sprintf("networkd/%d-%s.netdev.tpl", si, prefix)
	applyAndCleanUp(log, applier, tpl, src, dest, fileModeSystemd, false)

	src = mustTmpFile(prefix + "_network_")
	applier = newSystemdNetworkdApplier(src, data)
	dest = fmt.Sprintf("%s/%d-%s%s.network", systemdNetworkPath, di, prefix, suffix)
	tpl = fmt.Sprintf("networkd/%d-%s.network.tpl", si, prefix)
	applyAndCleanUp(log, applier, tpl, src, dest, fileModeSystemd, false)
}

func getEVPNIfaces(kb config) []EVPNIface {
	var result []EVPNIface

	vrfTableOffset := 1000
	for i, n := range kb.Networks {
		if n.Underlay != nil && *n.Underlay {
			continue
		}

		vrf := int(*n.Vrf)
		e := EVPNIface{}
		e.Comment = versionHeader(kb.MachineUUID)
		e.SVI.Comment = fmt.Sprintf("# svi (networkid: %s)", *n.Networkid)
		e.SVI.VLANID = VLANOffset + i
		e.SVI.Addresses = addBitlen(n.Ips)
		e.VXLAN.Comment = fmt.Sprintf("# vxlan (networkid: %s)", *n.Networkid)
		e.VXLAN.ID = vrf
		e.VXLAN.TunnelIP = kb.getUnderlayNetwork().Ips[0]
		e.VRF.Comment = fmt.Sprintf("# vrf (networkid: %s)", *n.Networkid)
		e.VRF.ID = vrf
		e.VRF.Table = vrfTableOffset + i
		result = append(result, e)
	}

	return result
}
