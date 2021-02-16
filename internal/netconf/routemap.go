package netconf

import (
	"fmt"
	"sort"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"inet.af/netaddr"
)

type importRule struct {
	targetVRF              string
	importVRFs             []string
	importPrefixes         []netaddr.IPPrefix
	importPrefixesNoExport []netaddr.IPPrefix
}

func importRulesForNetwork(kb KnowledgeBase, network models.V1MachineNetwork) *importRule {
	vrfName := vrfNameOf(network)

	if network.Networktype == nil || *network.Networktype == mn.Underlay {
		return nil
	}
	i := importRule{
		targetVRF: vrfName,
	}
	privatePrimaryNet := kb.getPrivatePrimaryNetwork()

	externalNets := kb.GetNetworks(mn.External)
	privateSecondarSharedNets := kb.GetNetworks(mn.PrivateSecondaryShared)

	nt := *network.Networktype
	switch nt {
	case mn.PrivatePrimaryUnshared:
		fallthrough
	case mn.PrivatePrimaryShared:
		// reach out from private primary network into public networks
		i.importVRFs = vrfNamesOf(externalNets)
		i.importPrefixes = getDestinationPrefixes(externalNets)

		// reach out from private primary network into shared private networks
		i.importVRFs = append(i.importVRFs, vrfNamesOf(privateSecondarSharedNets)...)
		i.importPrefixes = append(i.importPrefixes, prefixesOfNetworks(privateSecondarSharedNets)...)
	case mn.PrivateSecondaryShared:
		// reach out from private shared networks into private primary network
		i.importVRFs = []string{vrfNameOf(privatePrimaryNet)}
		i.importPrefixes = concatPfxSlices(prefixesOfNetwork(privatePrimaryNet), prefixesOfNetwork(network))

		// import destination prefixes of dmz networks from external networks
		if len(network.Destinationprefixes) > 0 {
			for _, pfx := range network.Destinationprefixes {
				for _, e := range externalNets {
					importExternalNet := false
					for _, epfx := range e.Destinationprefixes {
						if pfx == epfx {
							importExternalNet = true
							i.importPrefixes = append(i.importPrefixes, netaddr.MustParseIPPrefix(pfx))
						}
					}
					if importExternalNet {
						i.importVRFs = append(i.importVRFs, vrfNameOf(e))
					}
				}
			}
		}
	case mn.External:
		// reach out from public into private and other public networks
		i.importVRFs = []string{vrfNameOf(privatePrimaryNet)}
		i.importPrefixes = prefixesOfNetwork(network)

		nets := []models.V1MachineNetwork{privatePrimaryNet}
		if containsDefaultRoute(network.Destinationprefixes) {
			for _, r := range privateSecondarSharedNets {
				if containsDefaultRoute(r.Destinationprefixes) {
					nets = append(nets, r)
				}
			}
		}
		i.importPrefixesNoExport = prefixesOfNetworks(nets)
	}

	return &i
}

func (i *importRule) prefixLists() []IPPrefixList {
	var result []IPPrefixList
	seed := IPPrefixListSeqSeed
	afs := []AddressFamily{AddressFamilyIPv4, AddressFamilyIPv6}
	for _, af := range afs {
		pfxList := prefixLists(i.importPrefixesNoExport, af, false, seed, i.targetVRF)
		result = append(result, pfxList...)
		seed = IPPrefixListSeqSeed + len(pfxList)
		result = append(result, prefixLists(i.importPrefixes, af, true, seed, i.targetVRF)...)
		seed = IPPrefixListSeqSeed
	}

	return result
}

func prefixLists(prefixes []netaddr.IPPrefix, af AddressFamily, isExported bool, seed int, vrf string) []IPPrefixList {
	var result []IPPrefixList
	for _, prefix := range prefixes {
		if af == AddressFamilyIPv4 && !prefix.IP.Is4() {
			continue
		}

		if af == AddressFamilyIPv6 && !prefix.IP.Is6() {
			continue
		}

		specs := buildIPPrefixListSpecs(seed, prefix)
		for _, spec := range specs {
			name := namePrefixList(vrf, prefix, isExported)
			prefixList := IPPrefixList{
				Name:          name,
				Spec:          spec,
				AddressFamily: af,
			}
			result = append(result, prefixList)
		}
		seed++
	}
	return result
}

func concatPfxSlices(pfxSlices ...[]netaddr.IPPrefix) []netaddr.IPPrefix {
	res := []netaddr.IPPrefix{}
	for _, pfxSlice := range pfxSlices {
		res = append(res, pfxSlice...)
	}
	return res
}

func stringSliceToIPPrefix(s []string) []netaddr.IPPrefix {
	var result []netaddr.IPPrefix
	for _, e := range s {
		ipp, err := netaddr.ParseIPPrefix(e)
		if err != nil {
			continue
		}
		result = append(result, ipp)
	}
	return result
}

func getDestinationPrefixes(networks []models.V1MachineNetwork) []netaddr.IPPrefix {
	var result []netaddr.IPPrefix
	for _, network := range networks {
		result = append(result, stringSliceToIPPrefix(network.Destinationprefixes)...)
	}
	return result
}

func prefixesOfNetworks(networks []models.V1MachineNetwork) []netaddr.IPPrefix {
	var result []netaddr.IPPrefix
	for _, network := range networks {
		result = append(result, prefixesOfNetwork(network)...)
	}
	return result
}

func prefixesOfNetwork(network models.V1MachineNetwork) []netaddr.IPPrefix {
	return stringSliceToIPPrefix(network.Prefixes)
}

func vrfNameOf(n models.V1MachineNetwork) string {
	return fmt.Sprintf("vrf%d", *n.Vrf)
}

func vrfNamesOf(networks []models.V1MachineNetwork) []string {
	var result []string
	for _, n := range networks {
		result = append(result, vrfNameOf(n))
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

func (i *importRule) routeMaps() []RouteMap {
	var result []RouteMap

	order := RouteMapOrderSeed
	byName := byName(i.prefixLists())

	names := []string{}
	for n := range byName {
		names = append(names, n)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	for _, n := range names {
		prefixList := byName[n]
		match := fmt.Sprintf("match %s address prefix-list %s", prefixList.AddressFamily, n)
		entries := []string{match}
		if strings.HasSuffix(n, IPPrefixListNoExportSuffix) {
			entries = append(entries, "set community additive no-export")
		}

		routeMap := RouteMap{
			Name:    routeMapName(i.targetVRF),
			Policy:  Permit.String(),
			Order:   order,
			Entries: entries,
		}
		order += RouteMapOrderSeed

		result = append(result, routeMap)
	}

	routeMap := RouteMap{
		Name:   routeMapName(i.targetVRF),
		Policy: Deny.String(),
		Order:  order,
	}

	result = append(result, routeMap)

	return result
}

func routeMapName(vrfName string) string {
	return vrfName + "-import-map"
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

func namePrefixList(vrfName string, prefix netaddr.IPPrefix, isExported bool) string {
	af := ""
	if prefix.IP.Is6() {
		af = "-ipv6"
	}
	export := ""
	if !isExported {
		export = IPPrefixListNoExportSuffix
	}

	return fmt.Sprintf("%s-import-prefixes%s%s", vrfName, af, export)
}
