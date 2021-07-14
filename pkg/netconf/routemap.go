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
	Prefix    netaddr.IPPrefix
	Policy    AccessPolicy
	SourceVRF string
}

type importRule struct {
	TargetVRF              string
	ImportVRFs             []string
	ImportPrefixes         []importPrefix
	ImportPrefixesNoExport []importPrefix
}

func importRulesForNetwork(kb KnowledgeBase, network models.V1MachineNetwork) *importRule {
	vrfName := vrfNameOf(network)

	if network.Networktype == nil || *network.Networktype == mn.Underlay {
		return nil
	}
	i := importRule{
		TargetVRF: vrfName,
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
		i.ImportVRFs = vrfNamesOf(externalNets)
		i.ImportPrefixes = getDestinationPrefixes(externalNets)

		// deny public address of default network
		defaultNet := kb.GetDefaultRouteNetwork()
		if ip, err := netaddr.ParseIP(defaultNet.Ips[0]); err == nil {
			var bl uint8 = 32
			if ip.Is6() {
				bl = 128
			}
			i.ImportPrefixes = append(i.ImportPrefixes, importPrefix{
				Prefix:    netaddr.IPPrefixFrom(ip, bl),
				Policy:    Deny,
				SourceVRF: vrfNameOf(*defaultNet),
			})
		}

		// permit external routes
		i.ImportPrefixes = append(i.ImportPrefixes, prefixesOfNetworks(externalNets)...)

		// reach out from private network into shared private networks
		i.ImportVRFs = append(i.ImportVRFs, vrfNamesOf(privateSecondarySharedNets)...)
		i.ImportPrefixes = append(i.ImportPrefixes, prefixesOfNetworks(privateSecondarySharedNets)...)

		// reach out from private network to destination prefixes of private secondays shared networks
		for _, n := range privateSecondarySharedNets {
			for _, pfx := range n.Destinationprefixes {
				ppfx := netaddr.MustParseIPPrefix(pfx)
				isThere := false
				for _, i := range i.ImportPrefixes {
					if i.Prefix == ppfx {
						isThere = true
					}
				}
				if !isThere {
					i.ImportPrefixes = append(i.ImportPrefixes, importPrefix{
						Prefix:    ppfx,
						Policy:    Permit,
						SourceVRF: vrfNameOf(n),
					})
				}
			}
		}
	case mn.PrivateSecondaryShared:
		// reach out from private shared networks into private primary network
		i.ImportVRFs = []string{vrfNameOf(privatePrimaryNet)}
		i.ImportPrefixes = concatPfxSlices(prefixesOfNetwork(privatePrimaryNet), prefixesOfNetwork(network))

		// import destination prefixes of dmz networks from external networks
		if len(network.Destinationprefixes) > 0 {
			for _, pfx := range network.Destinationprefixes {
				for _, e := range externalNets {
					importExternalNet := false
					for _, epfx := range e.Destinationprefixes {
						if pfx == epfx {
							importExternalNet = true
							i.ImportPrefixes = append(i.ImportPrefixes, importPrefix{
								Prefix:    netaddr.MustParseIPPrefix(pfx),
								Policy:    Permit,
								SourceVRF: vrfNameOf(e),
							})
						}
					}
					if importExternalNet {
						i.ImportVRFs = append(i.ImportVRFs, vrfNameOf(e))
						i.ImportPrefixes = append(i.ImportPrefixes, prefixesOfNetwork(e)...)
					}
				}
			}
		}
	case mn.External:
		// reach out from public into private and other public networks
		i.ImportVRFs = []string{vrfNameOf(privatePrimaryNet)}
		i.ImportPrefixes = prefixesOfNetwork(network)

		nets := []models.V1MachineNetwork{privatePrimaryNet}

		if containsDefaultRoute(network.Destinationprefixes) {
			for _, r := range privateSecondarySharedNets {
				if containsDefaultRoute(r.Destinationprefixes) {
					nets = append(nets, r)
					i.ImportVRFs = append(i.ImportVRFs, vrfNameOf(r))
				}
			}
		}
		i.ImportPrefixesNoExport = prefixesOfNetworks(nets)
	}

	return &i
}

func (i *importRule) prefixLists() []IPPrefixList {
	var result []IPPrefixList
	seed := IPPrefixListSeqSeed
	afs := []AddressFamily{AddressFamilyIPv4, AddressFamilyIPv6}
	for _, af := range afs {
		pfxList := prefixLists(i.ImportPrefixesNoExport, af, false, seed, i.TargetVRF)
		result = append(result, pfxList...)

		seed = IPPrefixListSeqSeed + len(result)
		result = append(result, prefixLists(i.ImportPrefixes, af, true, seed, i.TargetVRF)...)
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
		if af == AddressFamilyIPv4 && !p.Prefix.IP().Is4() {
			continue
		}

		if af == AddressFamilyIPv6 && !p.Prefix.IP().Is6() {
			continue
		}

		specs := p.buildSpecs(seed)
		for _, spec := range specs {
			// self-importing prefixes is nonsense
			if vrf == p.SourceVRF {
				continue
			}
			name := p.name(vrf, isExported)
			prefixList := IPPrefixList{
				Name:          name,
				Spec:          spec,
				AddressFamily: af,
				SourceVRF:     p.SourceVRF,
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

func stringSliceToIPPrefix(s []string, sourceVrf string) []importPrefix {
	var result []importPrefix
	for _, e := range s {
		ipp, err := netaddr.ParseIPPrefix(e)
		if err != nil {
			continue
		}
		result = append(result, importPrefix{
			Prefix:    ipp,
			Policy:    Permit,
			SourceVRF: sourceVrf,
		})
	}
	return result
}

func getDestinationPrefixes(networks []models.V1MachineNetwork) []importPrefix {
	var result []importPrefix
	for _, network := range networks {
		result = append(result, stringSliceToIPPrefix(network.Destinationprefixes, vrfNameOf(network))...)
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
	return stringSliceToIPPrefix(network.Prefixes, vrfNameOf(network))
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

		matchVrf := fmt.Sprintf("match source-vrf %s", prefixList.SourceVRF)
		matchPfxList := fmt.Sprintf("match %s address prefix-list %s", prefixList.AddressFamily, n)
		entries := []string{matchVrf, matchPfxList}
		if strings.HasSuffix(n, IPPrefixListNoExportSuffix) {
			entries = append(entries, "set community additive no-export")
		}

		routeMap := RouteMap{
			Name:    routeMapName(i.TargetVRF),
			Policy:  Permit.String(),
			Order:   order,
			Entries: entries,
		}
		order += RouteMapOrderSeed

		result = append(result, routeMap)
	}

	routeMap := RouteMap{
		Name:   routeMapName(i.TargetVRF),
		Policy: Deny.String(),
		Order:  order,
	}

	result = append(result, routeMap)

	return result
}

func routeMapName(vrfName string) string {
	return vrfName + "-import-map"
}

func (i *importPrefix) buildSpecs(seq int) []string {
	var result []string
	var spec string

	if i.Prefix.Bits() == 0 {
		spec = fmt.Sprintf("%s %s", i.Policy, i.Prefix)

	} else {
		spec = fmt.Sprintf("seq %d %s %s le %d", seq, i.Policy, i.Prefix, i.Prefix.IP().BitLen())
	}

	result = append(result, spec)

	return result
}

func (i *importPrefix) name(targetVrf string, isExported bool) string {
	suffix := ""

	if i.Prefix.IP().Is6() {
		suffix = "-ipv6"
	}
	if !isExported {
		suffix += IPPrefixListNoExportSuffix
	}

	return fmt.Sprintf("%s-import-from-%s%s", targetVrf, i.SourceVRF, suffix)
}
