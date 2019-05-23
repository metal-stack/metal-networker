package main

import (
	"fmt"
	"os/exec"

	"go.uber.org/zap"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

// IfaceConfig represents a thing to apply changes to interfaces configuration.
type IfaceConfig struct {
	Applier network.Applier
	Log     zap.Logger
}

// IfacesData represents the information required to render interfaces configuration.
type IfacesData struct {
	Comment  string
	Underlay struct {
		Comment     string
		LoopbackIps []string
	}
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
}

// NewIfacesConfig constructs a new instance of this type.
func NewIfacesConfig(kb KnowledgeBase, tmpFile string) IfaceConfig {
	d := IfacesData{}
	d.Comment = fmt.Sprintf("# This file was auto generated for machine: '%s'.\n# Do not edit.", kb.Machineuuid)
	d.Underlay.Comment = getUnderlayComment(kb)
	d.Underlay.LoopbackIps = kb.mustGetUnderlay().Ips
	d.Bridge.Ports = getBridgePorts(kb)
	d.Bridge.Vids = getBridgeVLANIDs(kb)
	d.EVPNInterfaces = getEVPNInterfaces(kb)

	v := IfacesValidator{tmpFile}
	r := IfacesReloader{}
	a := network.NewNetworkApplier(d, v, r)
	return IfaceConfig{Applier: a}
}

// IfacesReloader can reload the service to apply changes to network interfaces.
type IfacesReloader struct {
}

// Reload reloads the service that applies changes to network interfaces.
func (r IfacesReloader) Reload() error {
	log.Info("running 'ifreload --all' to apply changes")
	return exec.Command("ifreload", "--all").Run()
}

// IfacesValidator can validate configuration for network interfaces.
type IfacesValidator struct {
	path string
}

// Validate validates network interfaces configuration.
func (v IfacesValidator) Validate() error {
	log.Info("running 'ifup --syntax-check --all --interfaces %s to validate changes.'", v.path)
	return exec.Command("ifup", "--syntax-check", "--all", "--interfaces", v.path).Run()
}

func getEVPNInterfaces(data KnowledgeBase) []EVPNIfaces {
	var result []EVPNIfaces
	for _, n := range data.Networks {
		if n.Underlay {
			continue
		}

		e := EVPNIfaces{}
		e.SVI.Comment = fmt.Sprintf("svi (networkid: %s)", n.Networkid)
		e.SVI.VlanID = n.Vlan
		e.SVI.Addresses = n.Ips

		e.VXLAN.Comment = fmt.Sprintf("vxlan (networkid: %s)", n.Networkid)
		e.VXLAN.ID = n.Vrf
		e.VXLAN.TunnelIP = data.mustGetUnderlay().Ips[0]

		e.VRF.Comment = fmt.Sprintf("vrf (networkid: %s)", n.Networkid)
		e.VRF.ID = n.Vrf

		result = append(result, e)
	}
	return result
}

func getBridgeVLANIDs(kb KnowledgeBase) string {
	result := ""
	for _, n := range kb.Networks {
		if n.Underlay {
			continue
		}
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
	for _, n := range kb.Networks {
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
	n := kb.mustGetUnderlay()
	return fmt.Sprintf("networkid: %s", n.Networkid)
}
