package netconf

import (
	"fmt"
	"sort"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"inet.af/netaddr"
)

type importPrefix struct {
	prefix netaddr.IPPrefix
	policy AccessPolicy
}

type importRule struct {
	targetVRF              string
	importVRFs             []string
	importPrefixes         []importPrefix
	importPrefixesNoExport []importPrefix
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
	privateSecondarySharedNets := kb.GetNetworks(mn.PrivateSecondaryShared)

	nt := *network.Networktype
	switch nt {
	case mn.PrivatePrimaryUnshared:
		fallthrough
	case mn.PrivatePrimaryShared:
		// reach out from private network into public networks
		i.importVRFs = vrfNamesOf(externalNets)
		i.importPrefixes = getDestinationPrefixes(externalNets)

		// deny public address of default network
		defaultNet := kb.getDefaultRouteNetwork()
		if ip, err := netaddr.ParseIP(defaultNet.Ips[0]); err == nil {
			var bl uint8 = 32
			if ip.Is6() {
				bl = 128
			}
			i.importPrefixes = append(i.importPrefixes, importPrefix{
				prefix: netaddr.IPPrefix{
					IP:   ip,
					Bits: bl,
				},
				policy: Deny,
			})
		}

		// permit external routes
		i.importPrefixes = append(i.importPrefixes, prefixesOfNetworks(externalNets)...)

		// reach out from private network into shared private networks
		i.importVRFs = append(i.importVRFs, vrfNamesOf(privateSecondarySharedNets)...)
		i.importPrefixes = append(i.importPrefixes, prefixesOfNetworks(privateSecondarySharedNets)...)

		// reach out from private network to destination prefixes of private secondays shared networks
		for _, n := range privateSecondarySharedNets {
			for _, pfx := range n.Destinationprefixes {
				ppfx := netaddr.MustParseIPPrefix(pfx)
				isThere := false
				for _, i := range i.importPrefixes {
					if i.prefix == ppfx {
						isThere = true
					}
				}
				if !isThere {
					i.importPrefixes = append(i.importPrefixes, importPrefix{
						prefix: ppfx,
						policy: Permit,
					})
				}
			}
		}
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
							i.importPrefixes = append(i.importPrefixes, importPrefix{
								prefix: netaddr.MustParseIPPrefix(pfx),
								policy: Permit,
							})
						}
					}
					if importExternalNet {
						i.importVRFs = append(i.importVRFs, vrfNameOf(e))
						i.importPrefixes = append(i.importPrefixes, prefixesOfNetwork(e)...)
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
			for _, r := range privateSecondarySharedNets {
				if containsDefaultRoute(r.Destinationprefixes) {
					nets = append(nets, r)
					i.importVRFs = append(i.importVRFs, vrfNameOf(r))
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

		seed = IPPrefixListSeqSeed + len(result)
		result = append(result, prefixLists(i.importPrefixes, af, true, seed, i.targetVRF)...)
	}

	return result
}

func prefixLists(
	prefixes []importPrefix,
	af AddressFamily,
	isExported bool,
	seed int,
	vrf string,
) []IPPrefixList {
	var result []IPPrefixList
	for _, p := range prefixes {
		if af == AddressFamilyIPv4 && !p.prefix.IP.Is4() {
			continue
		}

		if af == AddressFamilyIPv6 && !p.prefix.IP.Is6() {
			continue
		}

		specs := buildIPPrefixListSpecs(seed, p.policy, p.prefix)
		for _, spec := range specs {
			name := namePrefixList(vrf, p.prefix, isExported)
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

func concatPfxSlices(pfxSlices ...[]importPrefix) []importPrefix {
	res := []importPrefix{}
	for _, pfxSlice := range pfxSlices {
		res = append(res, pfxSlice...)
	}
	return res
}

func stringSliceToIPPrefix(s []string) []importPrefix {
	var result []importPrefix
	for _, e := range s {
		ipp, err := netaddr.ParseIPPrefix(e)
		if err != nil {
			continue
		}
		result = append(result, importPrefix{
			prefix: ipp,
			policy: Permit,
		})
	}
	return result
}

func getDestinationPrefixes(networks []models.V1MachineNetwork) []importPrefix {
	var result []importPrefix
	for _, network := range networks {
		result = append(result, stringSliceToIPPrefix(network.Destinationprefixes)...)
	}
	return result
}

func prefixesOfNetworks(networks []models.V1MachineNetwork) []importPrefix {
	var result []importPrefix
	for _, network := range networks {
		result = append(result, prefixesOfNetwork(network)...)
	}
	return result
}

func prefixesOfNetwork(network models.V1MachineNetwork) []importPrefix {
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

func buildIPPrefixListSpecs(seq int, policy AccessPolicy, prefix netaddr.IPPrefix) []string {
	var result []string
	var spec string

	if prefix.Bits == 0 {
		spec = fmt.Sprintf("%s %s", policy, prefix)

	} else {
		spec = fmt.Sprintf("seq %d %s %s le %d", seq, policy, prefix, prefix.IP.BitLen())
	}

	result = append(result, spec)

	return result
}

func namePrefixList(vrfName string, prefix netaddr.IPPrefix, isExported bool) string {
	suffix := ""

	if prefix.IP.Is6() {
		suffix = "-ipv6"
	}
	if !isExported {
		suffix += IPPrefixListNoExportSuffix
	}

	return fmt.Sprintf("%s-import-prefixes%s", vrfName, suffix)
}
