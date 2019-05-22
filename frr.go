package main

import (
	"fmt"
	"net"

	"git.f-i-ts.de/cloud-native/metallib/network"

	"os/exec"
)

const FRRVersion = "7.0"
const DefaultRoute = "0.0.0.0/0"
const RouteLeakFmt = "ip route %s vrf%d nexthop-vrf vrf%d"

type FrrData struct {
	FRRVersion string
	ASN        int64
	Comment    string
	Hostname   string
	RouterID   string
	VRFs       []VRF
}

type VRF struct {
	ID                int
	VNI               int
	RouteLeaks        []string
	NetworksAnnounced []string
}

type FrrConfig struct {
	Applier network.Applier
}

func NewFrrConfig(kb KnowledgeBase, tmpFile string) FrrConfig {
	d := FrrData{}
	d.ASN = kb.mustGetUnderlay().Asn
	d.Comment = fmt.Sprintf("# This file was auto generated for machine: '%s'.\n# Do not edit.", kb.Machineuuid)
	d.FRRVersion = FRRVersion
	d.Hostname = kb.Hostname
	d.RouterID = kb.mustGetUnderlay().Ips[0]
	d.VRFs = getVRFs(kb)

	v := FrrValidator{tmpFile}
	r := network.NewDBusReloader("frr.service")
	a := network.NewNetworkApplier(d, v, r)

	return FrrConfig{a}
}

type FrrValidator struct {
	path string
}

func (v FrrValidator) Validate() error {
	vtysh := fmt.Sprintf("vtysh --dryrun --inputfile %s", v.path)
	return exec.Command("bash", "-c", vtysh, v.path).Run()
}

func getVRFs(kb KnowledgeBase) []VRF {
	var result []VRF
	for _, n := range kb.Networks {
		// VRF BGP Instances are configured for tenant network (primary) and all external networks
		// (non underlay) to enable traffic from tenant network into external networks and vice versa.
		if n.Underlay {
			continue
		}
		vrf := VRF{}
		vrf.ID = n.Vrf
		vrf.VNI = n.Vrf
		if n.Primary {
			// The primary vrf contains a static route leak into vrfs of external networks.
			// In addition to this the primary vrf announces a default route to ask clients to route all traffic destined
			// to external networks to here.
			vrf.RouteLeaks = getOutRouteLeaks(kb)
			vrf.NetworksAnnounced = []string{DefaultRoute}
		} else {
			// VRF BGP instances of external networks contain a route leak to return traffic to tenant servers.
			vrf.RouteLeaks = getInRouteLeaks(kb)
			// TODO: Add information of additionally configured /32 prefixes of tenant servers to install.yaml.
			// Non-primary network VRF BGP instances needs to announce the /32 IP addresses that have be
			// configured to the tenant servers loopback interface in addition to the primary network ip. Currently
			// this information is not part of the install.yaml input data.
			vrf.NetworksAnnounced = []string{}
		}
		result = append(result, vrf)
	}
	return result
}

// Returns route leaks that are meant to be added to the tenant vrf to enable outgoing traffic into external networks.
func getOutRouteLeaks(kb KnowledgeBase) []string {
	var result []string
	for _, n := range kb.Networks {
		// The primary and underlay networks are not targets to route external traffic to.
		if n.Primary || n.Underlay {
			continue
		}
		for _, d := range n.Destinationprefixes {
			rl := fmt.Sprintf(RouteLeakFmt, d, n.Vrf, n.Vrf)
			result = append(result, rl)
		}
	}
	return result
}

// Returns route  leaks that are meant to be added to the external network vrfs to route traffic into the tenant vrf.
func getInRouteLeaks(kb KnowledgeBase) []string {
	var result []string
	n := kb.mustGetPrimary()
	for _, p := range n.Prefixes {
		_, cidr, _ := net.ParseCIDR(p)
		for _, i := range n.Ips {
			ip := net.ParseIP(i)
			if cidr.Contains(ip) {
				rl := fmt.Sprintf(RouteLeakFmt, p, n.Vrf, n.Vrf)
				result = append(result, rl)
			}
		}
	}
	return result
}
