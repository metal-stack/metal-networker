package netconf

import (
	"fmt"

	"git.f-i-ts.de/cloud-native/metallib/network"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"
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
		Loopback struct {
			Comment string
			IPs     []string
		}
	}

	// MachineIfacesData contains attributes required to render network interfaces configuration of a bare metal
	// server that functions as 'machine'.
	MachineIfacesData struct {
		CommonIfacesData
		LocalBGPIfaceData LocalBGPIfaceData
	}

	// LocalBGPIfaceData contains attributes required to render network interfaces configuration for a local BGP peering.
	LocalBGPIfaceData struct {
		Comment string
		IP      string
	}

	// FirewallIfacesData contains attributes required to render network interfaces configuration of a bare metal
	// server that functions as 'firewall'.
	FirewallIfacesData struct {
		CommonIfacesData
		Bridge struct {
			Ports string
			Vids  string
		}
		EVPNInterfaces []EVPNIfaces
	}

	// EVPNIfaces represents the information required to render EVPN interfaces configuration.
	EVPNIfaces struct {
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
	}

	// IfacesValidator defines the base type of an interfaces validator.
	IfacesValidator struct {
		path string
	}
)

// NewIfacesConfigApplier constructs a new instance of this type.
func NewIfacesConfigApplier(kind BareMetalType, kb KnowledgeBase, tmpFile string) network.Applier {
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
		primary := kb.getPrimaryNetwork()
		common.Loopback.Comment = fmt.Sprintf("networkid: %s", primary.Networkid)
		// Ensure that the ips of the primary network are the first ips at the loopback interface.
		// The first lo IP is used within network communication and other systems depend on seeing the first primary ip.
		common.Loopback.IPs = append(primary.Ips, kb.CollectIPs(External)...)
		localBGP, err := getLocalBGPIfaceData(kb.getPrimaryNetwork())
		if err != nil {
			log.Fatalf("error finding bgp ip: %v", err)
		}
		data = MachineIfacesData{
			CommonIfacesData:  common,
			LocalBGPIfaceData: localBGP,
		}
	default:
		log.Fatalf("unknown configuratorType of configurator: %v", kind)
	}

	validator := IfacesValidator{path: tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
}

func getLocalBGPIfaceData(primary Network) (LocalBGPIfaceData, error) {
	var result LocalBGPIfaceData
	bgpIP, err := getLocalBGPIP(primary)
	if err != nil {
		return result, err
	}

	result = LocalBGPIfaceData{
		Comment: fmt.Sprintf("local dummy interface to allow for peering locally with machine"),
		IP:      bgpIP,
	}
	return result, nil
}

// Validate network interfaces configuration. Assumes ifupdown2 is available.
func (v IfacesValidator) Validate() error {
	log.Infof("running 'ifup --syntax-check --all --interfaces %s to validate changes.'", v.path)
	return exec.NewVerboseCmd("ifup", "--syntax-check", "--all", "--interfaces", v.path).Run()
}

func getEVPNInterfaces(kb KnowledgeBase) []EVPNIfaces {
	var result []EVPNIfaces
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
