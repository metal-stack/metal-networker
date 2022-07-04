package netconf

import (
	"fmt"
	"net/netip"
	"reflect"
	"testing"
)

type network struct {
	vrf          string
	prefixes     []importPrefix
	destinations []importPrefix
}

var (
	defaultRoute           = importPrefix{Prefix: netip.MustParsePrefix("0.0.0.0/0"), Policy: Permit, SourceVRF: inetVrf}
	defaultRoute6          = importPrefix{Prefix: netip.MustParsePrefix("::/0"), Policy: Permit, SourceVRF: inetVrf}
	defaultRouteFromDMZ    = importPrefix{Prefix: netip.MustParsePrefix("0.0.0.0/0"), Policy: Permit, SourceVRF: dmzVrf}
	externalVrf            = "vrf104010"
	externalNet            = importPrefix{Prefix: netip.MustParsePrefix("100.127.129.0/24"), Policy: Permit, SourceVRF: externalVrf}
	externalDestinationNet = importPrefix{Prefix: netip.MustParsePrefix("100.127.1.0/24"), Policy: Permit, SourceVRF: externalVrf}
	privateVrf             = "vrf3981"
	privateNet             = importPrefix{Prefix: netip.MustParsePrefix("10.0.16.0/22"), Policy: Permit, SourceVRF: privateVrf}
	privateNet6            = importPrefix{Prefix: netip.MustParsePrefix("2002::/64"), Policy: Permit, SourceVRF: privateVrf}
	sharedVrf              = "vrf3982"
	sharedNet              = importPrefix{Prefix: netip.MustParsePrefix("10.0.18.0/22"), Policy: Permit, SourceVRF: sharedVrf}
	dmzVrf                 = "vrf3983"
	dmzNet                 = importPrefix{Prefix: netip.MustParsePrefix("10.0.20.0/22"), Policy: Permit, SourceVRF: dmzVrf}
	inetVrf                = "vrf104009"
	inetNet1               = importPrefix{Prefix: netip.MustParsePrefix("185.1.2.0/24"), Policy: Permit, SourceVRF: inetVrf}
	inetNet2               = importPrefix{Prefix: netip.MustParsePrefix("185.27.0.0/22"), Policy: Permit, SourceVRF: inetVrf}
	inetNet6               = importPrefix{Prefix: netip.MustParsePrefix("2a02:c00:20::/45"), Policy: Permit, SourceVRF: inetVrf}
	publicDefaultNet       = importPrefix{Prefix: netip.MustParsePrefix("185.1.2.3/32"), Policy: Deny, SourceVRF: inetVrf}
	publicDefaultNet2      = importPrefix{Prefix: netip.MustParsePrefix("10.0.20.2/32"), Policy: Deny, SourceVRF: dmzVrf}
	publicDefaultNetIPv6   = importPrefix{Prefix: netip.MustParsePrefix("2a02:c00:20::1/128"), Policy: Deny, SourceVRF: inetVrf}

	private = network{
		vrf:      privateVrf,
		prefixes: []importPrefix{privateNet},
	}

	private6 = network{
		vrf:      privateVrf,
		prefixes: []importPrefix{privateNet6},
	}

	inet = network{
		vrf:          inetVrf,
		prefixes:     []importPrefix{inetNet1, inetNet2},
		destinations: []importPrefix{defaultRoute},
	}

	inet6 = network{
		vrf:          inetVrf,
		prefixes:     []importPrefix{inetNet6},
		destinations: []importPrefix{defaultRoute6},
	}

	external = network{
		vrf:          externalVrf,
		destinations: []importPrefix{externalDestinationNet},
		prefixes:     []importPrefix{externalNet},
	}

	shared = network{
		vrf:      sharedVrf,
		prefixes: []importPrefix{sharedNet},
	}

	dmz = network{
		vrf:          dmzVrf,
		prefixes:     []importPrefix{dmzNet},
		destinations: []importPrefix{defaultRouteFromDMZ},
	}
)

func leakFrom(pfxs []importPrefix, sourceVrf string) []importPrefix {
	r := []importPrefix{}
	for _, e := range pfxs {
		i := e
		i.SourceVRF = sourceVrf
		r = append(r, i)
	}
	return r
}

func Test_importRulesForNetwork(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]map[string]ImportSettings
	}{
		{
			name:  "standard firewall with private primary unshared network, private secondary shared network, internet and mpls",
			input: "testdata/firewall.yaml",
			want: map[string]map[string]ImportSettings{
				// The target VRF
				private.vrf: {
					// Imported VRFs with their restrictions
					inet.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes),
					},
					external.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(external.destinations, external.prefixes),
					},
					shared.vrf: ImportSettings{
						ImportPrefixes: shared.prefixes,
					},
				},
				shared.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(private.prefixes, leakFrom(shared.prefixes, private.vrf)),
					},
				},
				inet.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes:         leakFrom(inet.prefixes, private.vrf),
						ImportPrefixesNoExport: private.prefixes,
					},
				},
				external.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes:         leakFrom(external.prefixes, private.vrf),
						ImportPrefixesNoExport: private.prefixes,
					},
				},
			},
		},
		{
			name:  "firewall of a shared private network (shared/storage firewall)",
			input: "testdata/firewall_shared.yaml",
			want: map[string]map[string]ImportSettings{
				shared.vrf: {
					inet.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes),
					},
				},
				inet.vrf: {
					shared.vrf: ImportSettings{
						ImportPrefixes:         leakFrom(inet.prefixes, shared.vrf),
						ImportPrefixesNoExport: shared.prefixes,
					},
				},
			},
		},
		{
			name:  "firewall of a private network with dmz network and internet (dmz firewall)",
			input: "testdata/firewall_dmz.yaml",
			want: map[string]map[string]ImportSettings{
				private.vrf: {
					inet.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes),
					},
					dmz.vrf: ImportSettings{
						ImportPrefixes: dmz.prefixes,
					},
				},
				dmz.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(private.prefixes, leakFrom(dmz.prefixes, private.vrf)),
					},
					inet.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(inet.destinations, inet.prefixes),
					},
				},
				inet.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes:         leakFrom(inet.prefixes, private.vrf),
						ImportPrefixesNoExport: private.prefixes,
					},
					dmz.vrf: ImportSettings{
						ImportPrefixesNoExport: dmz.prefixes,
					},
				},
			},
		},
		{
			name:  "firewall of a private network with dmz network (dmz app firewall)",
			input: "testdata/firewall_dmz_app.yaml",
			want: map[string]map[string]ImportSettings{
				private.vrf: {
					dmz.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices([]importPrefix{publicDefaultNet2}, dmz.prefixes, dmz.destinations),
					},
				},
				dmz.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(private.prefixes, leakFrom(dmz.prefixes, private.vrf)),
					},
				},
			},
		},
		{
			name:  "firewall of a private network with dmz network and storage (dmz app firewall)",
			input: "testdata/firewall_dmz_app_storage.yaml",
			want: map[string]map[string]ImportSettings{
				private.vrf: {
					shared.vrf: ImportSettings{
						ImportPrefixes: shared.prefixes,
					},
					dmz.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices([]importPrefix{publicDefaultNet2}, dmz.prefixes, dmz.destinations),
					},
				},
				dmz.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(private.prefixes, leakFrom(dmz.prefixes, private.vrf)),
					},
				},
				shared.vrf: {
					private.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(private.prefixes, leakFrom(shared.prefixes, private.vrf)),
					},
				},
			},
		},
		{
			name:  "firewall with ipv6 private network and ipv6 internet network",
			input: "testdata/firewall_ipv6.yaml",
			want: map[string]map[string]ImportSettings{
				private6.vrf: {
					inet6.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(inet6.destinations, []importPrefix{publicDefaultNetIPv6}, inet6.prefixes),
					},
					external.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(external.destinations, external.prefixes),
					},
					shared.vrf: ImportSettings{
						ImportPrefixes: shared.prefixes,
					},
				},
				shared.vrf: {
					private6.vrf: ImportSettings{
						ImportPrefixes: concatPfxSlices(private6.prefixes, leakFrom(shared.prefixes, private6.vrf)),
					},
				},
				inet6.vrf: {
					private6.vrf: ImportSettings{
						ImportPrefixes:         leakFrom(inet6.prefixes, private6.vrf),
						ImportPrefixesNoExport: private6.prefixes,
					},
				},
				external.vrf: {
					private6.vrf: ImportSettings{
						ImportPrefixes:         leakFrom(external.prefixes, private6.vrf),
						ImportPrefixesNoExport: private6.prefixes,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			kb := NewKnowledgeBase(tt.input)
			err := kb.Validate(Firewall)
			if err != nil {
				t.Errorf("%s is not valid: %v", tt.input, err)
				return
			}
			for _, network := range kb.Networks {
				got := importRulesForNetwork(kb, network)
				if got == nil {
					continue
				}
				gotBySourceVrf := got.bySourceVrf()
				targetVrf := fmt.Sprintf("vrf%d", *network.Vrf)
				want := tt.want[targetVrf]

				if !reflect.DeepEqual(gotBySourceVrf, want) {
					t.Errorf("importRulesForNetwork() \ntargetVrf: %s \ng: %v, \nw: %v", targetVrf, gotBySourceVrf, want)
				}
			}
		})
	}
}
