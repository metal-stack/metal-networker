package main

import (
	"fmt"
	"os/exec"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

type IfaceConfig struct {
	Applier network.Applier
}

type InterfacesData struct {
	Comment  string
	Underlay struct {
		Comment     string
		LoopbackIps []string
	}
	Bridge struct {
		Ports string
		Vids  string
	}
	EVPNInterfaces []EVPNInterfaces
}

type EVPNInterfaces struct {
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

func NewIfacesConfig(kb KnowledgeBase, tmpFile string) IfaceConfig {
	d := InterfacesData{}
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

type IfacesReloader struct {
}

func (r IfacesReloader) Reload() error {
	return exec.Command("ifreload", "--all").Run()
}

type IfacesValidator struct {
	path string
}

func (v IfacesValidator) Validate() error {
	return exec.Command("ifup", "--syntax-check", "--all", "--interfaces", v.path).Run()
}

func getEVPNInterfaces(data KnowledgeBase) []EVPNInterfaces {
	var result []EVPNInterfaces
	for _, n := range data.Networks {
		if n.Underlay {
			continue
		}

		e := EVPNInterfaces{}
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
