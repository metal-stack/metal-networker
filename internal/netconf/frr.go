package netconf

import (
	"fmt"
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
	// IPPrefixListNoExportSuffix defines the suffix to use for private IP ranges that must not be exported.
	IPPrefixListNoExportSuffix = "-no-export"
	// RouteMapOrderSeed defines the initial value for route-map order.
	RouteMapOrderSeed = 10
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
		data = FirewallFRRData{
			CommonFRRData: newCommonFRRData(net, kb),
			VRFs:          assembleVRFs(kb),
		}
	case Machine:
		net := kb.getPrimaryNetwork()
		bgpIP, err := getLocalBGPIP(kb.getPrimaryNetwork())
		if err != nil {
			log.Fatalf("error finding bgp ip: %v", err)
		}
		data = MachineFRRData{
			CommonFRRData: newCommonFRRData(net, kb),
			LocalBGPIP:    bgpIP,
		}
	default:
		log.Fatalf("unknown kind of bare metal: %v", kind)
	}

	validator := FRRValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
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
		prefixLists := assembleIPPrefixListsFor(vrfName, prefixes, IPPrefixListSeqSeed, kb)
		vrf := VRF{
			Identity: Identity{
				ID: network.Vrf,
			},
			VNI:            network.Vrf,
			ImportVRFNames: vrfNamesOf(targets...),
			IPPrefixLists:  prefixLists,
			RouteMaps:      assembleRouteMapsFor(vrfName, prefixLists),
		}
		result = append(result, vrf)
	}
	return result
}

func uniqueNames(prefixLists []IPPrefixList) []string {
	var result []string
	uniqueNames := make(map[string]struct{})
	for _, prefixList := range prefixLists {
		if _, isPresent := uniqueNames[prefixList.Name]; isPresent {
			continue
		}
		uniqueNames[prefixList.Name] = struct{}{}
		result = append(result, prefixList.Name)
	}
	return result
}

func assembleRouteMapsFor(vrfName string, prefixLists []IPPrefixList) []RouteMap {
	var result []RouteMap
	order := RouteMapOrderSeed
	prefListNames := uniqueNames(prefixLists)
	for _, prefListName := range prefListNames {
		entries := []string{"match ip address prefix-list " + prefListName}
		if strings.HasSuffix(prefListName, IPPrefixListNoExportSuffix) {
			entries = append(entries, "set community additive no-export")
		}
		routeMap := RouteMap{
			Name:    routeMapName(vrfName),
			Policy:  Permit.String(),
			Order:   order,
			Entries: entries,
		}
		order += RouteMapOrderSeed
		result = append(result, routeMap)
	}
	routeMap := RouteMap{
		Name:   routeMapName(vrfName),
		Policy: Deny.String(),
		Order:  order,
	}
	result = append(result, routeMap)
	return result
}

func routeMapName(vrfName string) string {
	return vrfName + "-import-map"
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
	spec := fmt.Sprintf("seq %d %s %s", seq, Permit, prefix)
	if !strings.HasSuffix(prefix, "/0") {
		spec += " le 32"
	}
	result = append(result, spec)
	return result
}

func assembleIPPrefixListsFor(vrfName string, prefixes []string, seed int, kb KnowledgeBase) []IPPrefixList {
	var result []IPPrefixList
	primary := kb.getPrimaryNetwork()
	for _, prefix := range prefixes {
		if len(prefix) == 0 {
			continue
		}
		specs := buildIPPrefixListSpecs(seed, prefix)
		for _, spec := range specs {
			name := namePrefixList(vrfName, primary, prefix)
			prefixList := IPPrefixList{
				Name: name,
				Spec: spec,
			}
			result = append(result, prefixList)
		}
		seed += len(specs)
	}
	return result
}

func namePrefixList(vrfName string, primary Network, prefix string) string {
	name := vrfName + "-import-prefixes"
	for _, primaryPrefix := range primary.Prefixes {
		if primaryPrefix == prefix {
			// tenant primary network ip addresses must not be visible in the external VRFs to avoid blown up routing tables
			name += IPPrefixListNoExportSuffix
		}
	}
	return name
}
