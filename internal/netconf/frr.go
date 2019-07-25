package netconf

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

const (
	// FRRVersion holds a string that is used in the frr.conf to define the FRR version.
	FRRVersion = "7.0"
	// TplFirewallFRR defines the name of the template to render FRR configuration to a 'firewall'.
	TplFirewallFRR = "frr.firewall.tpl"
	// TplMachineFRR defines the name of the template to render FRR configuration to a 'machine'.
	TplMachineFRR = "frr.machine.tpl"
	// IPPrefixListSeqSeed specifies the initial value for prefix lists sequence number.
	IPPrefixListSeqSeed = 100
)

type (
	// CommonFRRData contains attributes that are common to FRR configuration of all kind of bare metal servers.
	CommonFRRData struct {
		ASN        int64
		Comment    string
		FRRVersion string
		Hostname   string
		RouterID   string
	}

	// MachineFRRData contains attributes required to render frr.conf of bare metal servers that function as 'machine'.
	MachineFRRData struct {
		CommonFRRData
		LocalBGPIP string
	}

	// FirewallFRRData contains attributes required to render frr.conf of bare metal servers that function as 'firewall'.
	FirewallFRRData struct {
		CommonFRRData
		VRFs []VRF
	}

	// VRF represents data required to render VRF information into frr.conf.
	VRF struct {
		ID             int
		VNI            int
		ImportVRFNames []string
		IPPrefixLists  []IPPrefixList
		RouteMap       RouteMap
	}

	// RouteMap represents a route-map to permit or deny routes.
	RouteMap struct {
		Name    string
		Entries []string
		Policy  string
		Order   int
	}

	// IPPrefixList represents 'ip prefix-list' filtering mechanism to be used in combination with route-maps.
	IPPrefixList struct {
		Name string
		Spec string
	}

	// FRRValidator validates the frr.conf to apply.
	FRRValidator struct {
		path string
	}
)

// NewFrrConfigApplier constructs a new Applier of the given type of Bare Metal.
func NewFrrConfigApplier(kind BareMetalType, kb KnowledgeBase, tmpFile string) network.Applier {
	var data interface{}

	switch kind {
	case Firewall:
		net := kb.getUnderlayNetwork()
		common := newCommonFRRData(net, kb)
		vrfs := assembleVRFs(kb)
		data = FirewallFRRData{CommonFRRData: common, VRFs: vrfs}
	case Machine:
		net := kb.getPrimaryNetwork()
		common := newCommonFRRData(net, kb)
		localBGPIP := getLocalBGPIP(kb)
		data = MachineFRRData{common, localBGPIP}
	default:
		log.Fatalf("unknown kind of bare metal: %v", kind)
	}

	validator := FRRValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
}

func getLocalBGPIP(kb KnowledgeBase) string {
	primaryIPs := kb.getPrimaryNetwork().Ips
	ip := net.ParseIP(primaryIPs[0])
	ip[len(ip)-1] = 0
	return ip.String()
}

func newCommonFRRData(net Network, kb KnowledgeBase) CommonFRRData {
	return CommonFRRData{FRRVersion: FRRVersion, Hostname: kb.Hostname, Comment: versionHeader(kb.Machineuuid),
		ASN: net.Asn, RouterID: net.Ips[0]}
}

// Validate can be used to run validation on FRR configuration using vtysh.
func (v FRRValidator) Validate() error {
	vtysh := fmt.Sprintf("vtysh --dryrun --inputfile %s", v.path)
	log.Infof("running '%s' to validate changes.'", vtysh)
	return exec.NewVerboseCmd("bash", "-c", vtysh, v.path).Run()
}

func getDestinationPrefixes(networks []Network) []string {
	var result []string
	for _, network := range networks {
		result = append(result, network.Destinationprefixes...)
	}
	return result
}

func getPrefixes(networks ...Network) []string {
	var result []string
	for _, network := range networks {
		result = append(result, network.Prefixes...)
	}
	return result
}

func assembleVRFs(kb KnowledgeBase) []VRF {
	var result []VRF
	networks := kb.GetNetworks(Primary, External)
	for _, network := range networks {
		var targets []Network
		var prefixes []string
		if network.Primary {
			// reach out from primary into to external networks
			targets = kb.GetNetworks(External)
			prefixes = getDestinationPrefixes(targets)
		} else {
			// reach out from external into primary and other external networks
			targets = kb.GetNetworks(Primary)
			prefixes = getPrefixes(append(targets, network)...)
		}
		vrfName := "vrf" + strconv.Itoa(network.Vrf)
		vrf := VRF{
			ID:             network.Vrf,
			VNI:            network.Vrf,
			ImportVRFNames: vrfNamesOf(targets...),
			IPPrefixLists:  assembleIPPrefixListsFor(vrfName, prefixes, IPPrefixListSeqSeed),
			RouteMap:       assembleRouteMapFor(vrfName),
		}
		result = append(result, vrf)
	}
	return result
}

func assembleRouteMapFor(vrfName string) RouteMap {
	var result RouteMap
	entries := []string{"match ip address prefix-list " + vrfName + "-import-prefixes"}
	result = RouteMap{
		Name:    vrfName + "-import-map",
		Policy:  "permit",
		Order:   10,
		Entries: entries,
	}
	return result
}

func vrfNamesOf(networks ...Network) []string {
	var result []string
	for _, n := range networks {
		vrf := fmt.Sprintf("vrf%d", n.Vrf)
		result = append(result, vrf)
	}
	return result
}

func buildIPPrefixListSpecs(seq int, prefix string) []string {
	var result []string
	spec := fmt.Sprintf("seq %d permit %s", seq, prefix)
	if !strings.HasSuffix(prefix, "/0") {
		spec += " le 32"
	}
	result = append(result, spec)
	return result
}

func assembleIPPrefixListsFor(vrfName string, prefixes []string, seed int) []IPPrefixList {
	var result []IPPrefixList
	for _, prefix := range prefixes {
		if len(prefix) == 0 {
			continue
		}
		specs := buildIPPrefixListSpecs(seed, prefix)
		for _, spec := range specs {
			prefixList := IPPrefixList{
				Name: vrfName + "-import-prefixes",
				Spec: spec,
			}
			result = append(result, prefixList)
		}
		seed += len(specs)
	}
	return result
}
