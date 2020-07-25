package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// TplFirewallIfaces defines the name of the template to render interfaces configuration for firewalls.
	TplFirewallIfaces = "interfaces.firewall.tpl"
	// TplMachineIfaces defines the name of the template to render interfaces configuration for machines.
	TplMachineIfaces = "lo.network.machine.tpl"
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

// NewIfacesConfigApplier constructs a new instance of this type.
func NewIfacesConfigApplier(kind BareMetalType, kb KnowledgeBase, tmpFile string) net.Applier {
	var data interface{}
	var validator net.Validator

	d := IfacesData{
		Comment: versionHeader(kb.Machineuuid),
	}

	d.Loopback.Comment = fmt.Sprintf("networkid: %s", kb.getUnderlayNetwork().Networkid)
	validator = SystemdNetworkdValidator{path: tmpFile}
	switch kind {
	case Firewall:
		d.Loopback.IPs = kb.getUnderlayNetwork().Ips
		d.Tenants = getTenants(kb)
	case Machine:
		private := kb.getPrivateNetwork()
		d.Loopback.Comment = fmt.Sprintf("networkid: %s", private.Networkid)
		// Ensure that the ips of the private network are the first ips at the loopback interface.
		// The first lo IP is used within network communication and other systems depend on seeing the first private ip.
		d.Loopback.IPs = append(private.Ips, kb.CollectIPs(Public)...)
	default:
		log.Fatalf("unknown configuratorType of configurator: %v", kind)
	}

	return net.NewNetworkApplier(data, validator, nil)
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
		e.SVI.Comment = fmt.Sprintf("svi (networkid: %s)", n.Networkid)
		e.SVI.VlanID = n.Vlan
		e.SVI.Addresses = n.Ips
		e.VXLAN.Comment = fmt.Sprintf("vxlan (networkid: %s)", n.Networkid)
		e.VXLAN.ID = n.Vrf
		e.VXLAN.TunnelIP = kb.getUnderlayNetwork().Ips[0]
		e.VRF.Comment = fmt.Sprintf("vrf (networkid: %s)", n.Networkid)
		e.VRF.ID = n.Vrf
		e.VRF.Table = offset + i
		result = append(result, e)
	}

	return result
}
