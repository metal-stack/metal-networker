package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/net"

	"github.com/metal-stack/metal-networker/pkg/exec"
)

const (
	// TplFirewallIfaces defines the name of the template to render interfaces configuration for firewalls.
	TplFirewallIfaces = "interfaces.firewall.tpl"
	// TplMachineIfaces defines the name of the template to render interfaces configuration for machines.
	TplMachineIfaces = "interfaces.machine.tpl"
)

type (
	// CommonIfacesData contains attributes required to render common network interfaces configuration of a bare metal
	// server.
	CommonIfacesData struct {
		Comment  string
		Loopback Loopback
	}

	// MachineIfacesData contains attributes required to render network interfaces configuration of a bare metal
	// server that functions as 'machine'.
	MachineIfacesData struct {
		CommonIfacesData
	}

	// FirewallIfacesData contains attributes required to render network interfaces configuration of a bare metal
	// server that functions as 'firewall'.
	FirewallIfacesData struct {
		CommonIfacesData
		Bridge         Bridge
		EVPNInterfaces []EVPNIface
	}

	// IfacesValidator defines the base type of an interfaces validator.
	IfacesValidator struct {
		path string
	}
)

// NewIfacesConfigApplier constructs a new instance of this type.
func NewIfacesConfigApplier(kind BareMetalType, kb KnowledgeBase, tmpFile string) net.Applier {
	var data interface{}

	common := CommonIfacesData{
		Comment: versionHeader(kb.Machineuuid),
	}

	switch kind {
	case Firewall:
		common.Loopback.Comment = fmt.Sprintf("networkid: %s", kb.getUnderlayNetwork().Networkid)
		common.Loopback.IPs = kb.getUnderlayNetwork().Ips
		f := FirewallIfacesData{}
		f.CommonIfacesData = common
		f.Bridge.Ports = getBridgePorts(kb)
		f.Bridge.Vids = getBridgeVLANIDs(kb)
		f.EVPNInterfaces = getEVPNInterfaces(kb)
		data = f
	case Machine:
		private := kb.getPrivateNetwork()
		common.Loopback.Comment = fmt.Sprintf("networkid: %s", private.Networkid)
		// Ensure that the ips of the private network are the first ips at the loopback interface.
		// The first lo IP is used within network communication and other systems depend on seeing the first private ip.
		common.Loopback.IPs = append(private.Ips, kb.CollectIPs(Public)...)
		data = MachineIfacesData{
			CommonIfacesData: common,
		}
	default:
		log.Fatalf("unknown configuratorType of configurator: %v", kind)
	}

	validator := IfacesValidator{path: tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// Validate network interfaces configuration. Assumes ifupdown2 is available.
func (v IfacesValidator) Validate() error {
	log.Infof("running 'ifup --syntax-check --all --interfaces %s to validate changes.'", v.path)
	return exec.NewVerboseCmd("ifup", "--syntax-check", "--all", "--interfaces", v.path).Run()
}

func getEVPNInterfaces(kb KnowledgeBase) []EVPNIface {
	var result []EVPNIface

	for _, n := range kb.Networks {
		if n.Underlay {
			continue
		}

		e := EVPNIface{}
		e.SVI.Comment = fmt.Sprintf("svi (networkid: %s)", n.Networkid)
		e.SVI.VlanID = n.Vlan
		e.SVI.Addresses = n.Ips
		e.VXLAN.Comment = fmt.Sprintf("vxlan (networkid: %s)", n.Networkid)
		e.VXLAN.ID = n.Vrf
		e.VXLAN.TunnelIP = kb.getUnderlayNetwork().Ips[0]
		e.VRF.Comment = fmt.Sprintf("vrf (networkid: %s)", n.Networkid)
		e.VRF.ID = n.Vrf
		result = append(result, e)
	}

	return result
}

func getBridgeVLANIDs(kb KnowledgeBase) string {
	result := ""
	networks := kb.GetNetworks(Private, Public)

	for _, n := range networks {
		if result == "" {
			result = fmt.Sprintf("%d", n.Vlan)
		} else {
			result = fmt.Sprintf("%s %d", result, n.Vlan)
		}
	}

	return result
}

func getBridgePorts(kb KnowledgeBase) string {
	result := ""
	networks := kb.GetNetworks(Private, Public)

	for _, n := range networks {
		if n.Underlay {
			continue
		}

		if result == "" {
			result = fmt.Sprintf("vni%d", n.Vrf)
		} else {
			result = fmt.Sprintf("%s vni%d", result, n.Vrf)
		}
	}

	return result
}
