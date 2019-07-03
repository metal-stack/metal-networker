package netconf

import (
	"fmt"

	"git.f-i-ts.de/cloud-native/metallib/network"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"
)

// TplFirewallIfaces defines the name of the template to render interfaces configuration.
const (
	TplFirewallIfaces = "interfaces.firewall.tpl"
	TplMachineIfaces  = "interfaces.machine.tpl"
)

// CommonIfacesData contains attributes required to render common network interfaces configuration of a bare metal
// server.
type CommonIfacesData struct {
	Comment  string
	Underlay struct {
		Comment     string
		LoopbackIps []string
	}
}

// MachineIfacesData contains attributes required to render network interfaces configuration of a bare metal
// server that functions as 'machine'.
type MachineIfacesData struct {
	CommonIfacesData
}

// FirewallIfacesData contains attributes required to render network interfaces configuration of a bare metal
// server that functions as 'firewall'.
type FirewallIfacesData struct {
	CommonIfacesData
	Bridge struct {
		Ports string
		Vids  string
	}
	EVPNInterfaces []EVPNIfaces
}

// EVPNIfaces represents the information required to render EVPN interfaces configuration.
type EVPNIfaces struct {
	VRF struct {
		ID      int
		Comment string
	}
	SVI struct {
		VlanID    int
		Comment   string
		Addresses []string
	}
	VXLAN struct {
		Comment  string
		ID       int
		TunnelIP string
	}
	PostUpCommands  []string
	PreDownCommands []string
}

// CommonIfacesValidator defines the base type of an interfaces validator.
type CommonIfacesValidator struct {
	path string
}

// MachineIfacesValidator defines a type to validate interfaces configuration of a bare metal server that function as
// 'machine'.
type MachineIfacesValidator struct {
	CommonIfacesValidator
}

// FirewallIfacesValidator defines a type to validate interfaces configuration of a bare metal server that function as
// 'firewall'.
type FirewallIfacesValidator struct {
	CommonIfacesValidator
}

// NewIfacesConfigApplier constructs a new instance of this type.
func NewIfacesConfigApplier(kind BareMetalType, kb KnowledgeBase, tmpFile string) network.Applier {
	var data interface{}
	var validator network.Validator

	switch kind {
	case Firewall:
		common := CommonIfacesData{}
		common.Comment = versionHeader(kb.Machineuuid)
		common.Underlay.Comment = getUnderlayComment(kb)
		common.Underlay.LoopbackIps = kb.getUnderlayNetwork().Ips

		f := FirewallIfacesData{}
		f.CommonIfacesData = common
		f.Bridge.Ports = getBridgePorts(kb)
		f.Bridge.Vids = getBridgeVLANIDs(kb)
		f.EVPNInterfaces = getEVPNInterfaces(kb)

		data = f
		validator = FirewallIfacesValidator{CommonIfacesValidator{path: tmpFile}}
	case Machine:
		common := CommonIfacesData{}
		common.Comment = versionHeader(kb.Machineuuid)
		common.Underlay.Comment = getUnderlayComment(kb)
		common.Underlay.LoopbackIps = kb.getPrimaryNetwork().Ips

		data = MachineIfacesData{common}
		validator = MachineIfacesValidator{CommonIfacesValidator{path: tmpFile}}
	default:
		log.Fatalf("unknown configuratorType of configurator: %v", kind)
	}

	return network.NewNetworkApplier(data, validator, nil)
}

// Validate 'machine' network interfaces configuration.
func (v MachineIfacesValidator) Validate() error {
	log.Infof("running 'ifup --no-act --all --interfaces %s' to validate changes", v.path)
	return exec.NewVerboseCmd("ifup", "--no-act", "--all", "--interfaces", v.path).Run()
}

// Validate 'firewall' network interfaces configuration.
func (v FirewallIfacesValidator) Validate() error {
	log.Infof("running 'ifup --syntax-check --all --interfaces %s to validate changes.'", v.path)
	return exec.NewVerboseCmd("ifup", "--syntax-check", "--all", "--interfaces", v.path).Run()
}

func getEVPNInterfaces(kb KnowledgeBase) []EVPNIfaces {
	var result []EVPNIfaces
	primary := kb.getPrimaryNetwork()
	for _, n := range kb.Networks {
		if n.Underlay {
			continue
		}

		e := EVPNIfaces{}
		e.SVI.Comment = fmt.Sprintf("svi (networkid: %s)", n.Networkid)
		e.SVI.VlanID = n.Vlan
		e.SVI.Addresses = n.Ips

		e.VXLAN.Comment = fmt.Sprintf("vxlan (networkid: %s)", n.Networkid)
		e.VXLAN.ID = n.Vrf
		e.VXLAN.TunnelIP = kb.getUnderlayNetwork().Ips[0]

		e.VRF.Comment = fmt.Sprintf("vrf (networkid: %s)", n.Networkid)
		e.VRF.ID = n.Vrf

		svi := fmt.Sprintf("vlan%d", n.Vrf)
		if n.Nat {
			for _, p := range primary.Prefixes {
				e.PostUpCommands = []string{fmt.Sprintf("iptables -t nat -A POSTROUTING -s %s -o %s -j MASQUERADE", p, svi)}
				e.PreDownCommands = []string{fmt.Sprintf("iptables -t nat -D POSTROUTING -s %s -o %s -j MASQUERADE", p, svi)}
			}
		}

		result = append(result, e)
	}
	return result
}

func getBridgeVLANIDs(kb KnowledgeBase) string {
	result := ""
	networks := kb.GetNetworks(Primary, External)
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
	networks := kb.GetNetworks(Primary, External)
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

func getUnderlayComment(kb KnowledgeBase) string {
	n := kb.getUnderlayNetwork()
	return fmt.Sprintf("networkid: %s", n.Networkid)
}
