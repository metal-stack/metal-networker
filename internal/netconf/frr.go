package netconf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"github.com/metal-stack/metal-networker/pkg/exec"
	"github.com/metal-stack/metal-networker/pkg/net"
	"inet.af/netaddr"
)

const (
	// FRRVersion holds a string that is used in the frr.conf to define the FRR version.
	FRRVersion = "7.5"
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
	// AddressFamilyIPv4 is the name for this address family for the routing daemon.
	AddressFamilyIPv4 = "ip"
	// AddressFamilyIPv6 is the name for this address family for the routing daemon.
	AddressFamilyIPv6 = "ipv6"
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

	// AddressFamily is the address family for the routing daemon.
	AddressFamily string
)

// NewFrrConfigApplier constructs a new Applier of the given type of Bare Metal.
func NewFrrConfigApplier(kind BareMetalType, kb KnowledgeBase, tmpFile string) net.Applier {
	var data interface{}

	switch kind {
	case Firewall:
		net := kb.getUnderlayNetwork()
		data = FirewallFRRData{
			CommonFRRData: CommonFRRData{
				FRRVersion: FRRVersion,
				Hostname:   kb.Hostname,
				Comment:    versionHeader(kb.Machineuuid),
				ASN:        *net.Asn,
				RouterID:   routerID(net),
			},
			VRFs: assembleVRFs(kb),
		}
	case Machine:
		net := kb.getPrivatePrimaryNetwork()
		data = MachineFRRData{
			CommonFRRData: CommonFRRData{
				FRRVersion: FRRVersion,
				Hostname:   kb.Hostname,
				Comment:    versionHeader(kb.Machineuuid),
				ASN:        *net.Asn,
				RouterID:   routerID(net),
			},
		}
	default:
		log.Fatalf("unknown kind of bare metal: %v", kind)
	}

	validator := FRRValidator{tmpFile}

	return net.NewNetworkApplier(data, validator, nil)
}

// routerID will calculate the bgp router-id which must only be specified in the ipv6 range.
// returns 0.0.0.0 for errornous ip addresses and 169.254.255.255 for ipv6
// TODO prepare machine allocations with ipv6 primary address and tests
func routerID(net models.V1MachineNetwork) string {
	if len(net.Ips) < 1 {
		return "0.0.0.0"
	}
	ip, err := netaddr.ParseIP(net.Ips[0])
	if err != nil {
		return "0.0.0.0"
	}
	if ip.Is4() {
		return ip.String()
	}
	return "169.254.255.255"
}

// Validate can be used to run validation on FRR configuration using vtysh.
func (v FRRValidator) Validate() error {
	vtysh := fmt.Sprintf("vtysh --dryrun --inputfile %s", v.path)
	log.Infof("running '%s' to validate changes.'", vtysh)

	return exec.NewVerboseCmd("bash", "-c", vtysh, v.path).Run()
}

func getDestinationPrefixes(networks []models.V1MachineNetwork) []string {
	var result []string
	for _, network := range networks {
		result = append(result, network.Destinationprefixes...)
	}

	return result
}

func getPrefixes(networks ...models.V1MachineNetwork) []string {
	var result []string
	for _, network := range networks {
		result = append(result, network.Prefixes...)
	}

	return result
}

func assembleVRFs(kb KnowledgeBase) []VRF {
	var result []VRF

	privatePrimary := kb.getPrivatePrimaryNetwork()
	networks := kb.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared, mn.PrivateSecondaryShared, mn.External)

	for _, network := range networks {
		var targets []models.V1MachineNetwork
		var prefixes []string

		if network.Networktype == nil {
			continue
		}
		nt := *network.Networktype
		switch nt {
		case mn.PrivatePrimaryUnshared:
			fallthrough
		case mn.PrivatePrimaryShared:
			// reach out from private primary network into public networks
			publicTargets := kb.GetNetworks(mn.External)
			prefixes = getDestinationPrefixes(publicTargets)
			targets = append(targets, publicTargets...)

			// reach out from private primary network into shared private networks
			privateSharedTargets := kb.GetNetworks(mn.PrivateSecondaryShared)
			prefixes = append(prefixes, getPrefixes(privateSharedTargets...)...)
			targets = append(targets, privateSharedTargets...)
		case mn.PrivateSecondaryShared:
			// reach out from private shared networks into private primary network
			targets = []models.V1MachineNetwork{privatePrimary}
			prefixes = getPrefixes(append(targets, network)...)
		case mn.External:
			// reach out from public into private and other public networks
			targets = []models.V1MachineNetwork{privatePrimary}
			prefixes = getPrefixes(append(targets, network)...)
		}
		shared := (nt == mn.PrivatePrimaryShared || nt == mn.PrivateSecondaryShared)
		vrfID := network.Vrf
		vrfName := "vrf" + strconv.Itoa(int(*vrfID))

		prefixLists := assembleIPPrefixListsFor(vrfName, prefixes, IPPrefixListSeqSeed, kb, shared, AddressFamilyIPv4)
		prefixLists = append(prefixLists, assembleIPPrefixListsFor(vrfName, prefixes, IPPrefixListSeqSeed, kb, shared, AddressFamilyIPv6)...)
		vrf := VRF{
			Identity: Identity{
				ID: int(*network.Vrf),
			},
			VNI:            int(*network.Vrf),
			ImportVRFNames: vrfNamesOf(targets...),
			IPPrefixLists:  prefixLists,
			RouteMaps:      assembleRouteMapsFor(vrfName, prefixLists),
		}
		result = append(result, vrf)
	}

	return result
}

func byName(prefixLists []IPPrefixList) map[string]IPPrefixList {
	byName := map[string]IPPrefixList{}
	for _, prefixList := range prefixLists {
		if _, isPresent := byName[prefixList.Name]; isPresent {
			continue
		}

		byName[prefixList.Name] = prefixList
	}

	return byName
}

func assembleRouteMapsFor(vrfName string, prefixLists []IPPrefixList) []RouteMap {
	var result []RouteMap

	order := RouteMapOrderSeed
	byName := byName(prefixLists)

	for prefListName, prefixList := range byName {
		match := fmt.Sprintf("match %s address prefix-list %s", prefixList.AddressFamily, prefListName)
		entries := []string{match}
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

func vrfNamesOf(networks ...models.V1MachineNetwork) []string {
	var result []string

	for _, n := range networks {
		vrf := fmt.Sprintf("vrf%d", *n.Vrf)
		result = append(result, vrf)
	}

	return result
}

func buildIPPrefixListSpecs(seq int, prefix netaddr.IPPrefix) []string {
	var result []string

	spec := fmt.Sprintf("seq %d %s %s", seq, Permit, prefix)
	if prefix.Bits != 0 {
		spec += fmt.Sprintf(" le %d", prefix.IP.BitLen())
	}

	result = append(result, spec)

	return result
}

func assembleIPPrefixListsFor(vrfName string, prefixes []string, seed int, kb KnowledgeBase, shared bool, af AddressFamily) []IPPrefixList {
	var result []IPPrefixList

	private := kb.getPrivatePrimaryNetwork()

	for _, prefix := range prefixes {
		if len(prefix) == 0 {
			continue
		}

		p, err := netaddr.ParseIPPrefix(prefix)
		if err != nil {
			continue
		}

		if af == AddressFamilyIPv4 && !p.IP.Is4() {
			continue
		}

		if af == AddressFamilyIPv6 && !p.IP.Is6() {
			continue
		}

		specs := buildIPPrefixListSpecs(seed, p)

		for _, spec := range specs {
			name := namePrefixList(vrfName, private, p, shared, af)
			prefixList := IPPrefixList{
				Name:          name,
				Spec:          spec,
				AddressFamily: af,
			}
			result = append(result, prefixList)
		}

		seed += len(specs)
	}

	return result
}

func namePrefixList(vrfName string, private models.V1MachineNetwork, prefix netaddr.IPPrefix, shared bool, af AddressFamily) string {
	name := fmt.Sprintf("%s-import-prefixes", vrfName)
	if af != AddressFamilyIPv4 {
		name += fmt.Sprintf("-%s", af)
	}

	for _, privatePrefix := range private.Prefixes {
		if privatePrefix == prefix.String() && !shared {
			// tenant private network ip addresses must not be visible in the public VRFs to avoid blown up routing tables
			name += IPPrefixListNoExportSuffix
		}
	}

	return name
}
