package netconf

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"inet.af/netaddr"
)

type network struct {
	vrf          string
	prefixes     []importPrefix
	destinations []importPrefix
}

var (
	defaultRoute           = importPrefix{Prefix: netaddr.MustParseIPPrefix("0.0.0.0/0"), Policy: Permit, SourceVRF: inetVrf}
	defaultRoute6          = importPrefix{Prefix: netaddr.MustParseIPPrefix("::/0"), Policy: Permit, SourceVRF: inetVrf}
	defaultRouteFromDMZ    = importPrefix{Prefix: netaddr.MustParseIPPrefix("0.0.0.0/0"), Policy: Permit, SourceVRF: dmzVrf}
	externalVrf            = "vrf104010"
	externalNet            = importPrefix{Prefix: netaddr.MustParseIPPrefix("100.127.129.0/24"), Policy: Permit, SourceVRF: externalVrf}
	externalDestinationNet = importPrefix{Prefix: netaddr.MustParseIPPrefix("100.127.1.0/24"), Policy: Permit, SourceVRF: externalVrf}
	privateVrf             = "vrf3981"
	privateNet             = importPrefix{Prefix: netaddr.MustParseIPPrefix("10.0.16.0/22"), Policy: Permit, SourceVRF: privateVrf}
	privateNet6            = importPrefix{Prefix: netaddr.MustParseIPPrefix("2002::/64"), Policy: Permit, SourceVRF: privateVrf}
	sharedVrf              = "vrf3982"
	sharedNet              = importPrefix{Prefix: netaddr.MustParseIPPrefix("10.0.18.0/22"), Policy: Permit, SourceVRF: sharedVrf}
	dmzVrf                 = "vrf3983"
	dmzNet                 = importPrefix{Prefix: netaddr.MustParseIPPrefix("10.0.20.0/22"), Policy: Permit, SourceVRF: dmzVrf}
	inetVrf                = "vrf104009"
	inetNet1               = importPrefix{Prefix: netaddr.MustParseIPPrefix("185.1.2.0/24"), Policy: Permit, SourceVRF: inetVrf}
	inetNet2               = importPrefix{Prefix: netaddr.MustParseIPPrefix("185.27.0.0/22"), Policy: Permit, SourceVRF: inetVrf}
	inetNet6               = importPrefix{Prefix: netaddr.MustParseIPPrefix("2a02:c00:20::/45"), Policy: Permit, SourceVRF: inetVrf}
	publicDefaultNet       = importPrefix{Prefix: netaddr.MustParseIPPrefix("185.1.2.3/32"), Policy: Deny, SourceVRF: inetVrf}
	publicDefaultNet2      = importPrefix{Prefix: netaddr.MustParseIPPrefix("10.0.20.2/32"), Policy: Deny, SourceVRF: dmzVrf}
	publicDefaultNetIPv6   = importPrefix{Prefix: netaddr.MustParseIPPrefix("2a02:c00:20::1/128"), Policy: Deny, SourceVRF: inetVrf}

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

func Test_importRulesForNetwork(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*importRule
	}{
		{
			name:  "standard firewall with private primary unshared network, private secondary shared network, internet and mpls",
			input: "testdata/firewall.yaml",
			want: []*importRule{
				{
					TargetVRF:      private.vrf,
					ImportVRFs:     []string{inet.vrf, external.vrf, shared.vrf},
					ImportPrefixes: concatPfxSlices(inet.destinations, external.destinations, []importPrefix{publicDefaultNet}, inet.prefixes, external.prefixes, shared.prefixes),
				},
				{
					TargetVRF:      shared.vrf,
					ImportVRFs:     []string{private.vrf},
					ImportPrefixes: concatPfxSlices(private.prefixes, shared.prefixes),
				},
				{
					TargetVRF:              inet.vrf,
					ImportVRFs:             []string{private.vrf},
					ImportPrefixes:         inet.prefixes,
					ImportPrefixesNoExport: private.prefixes,
				},
				nil,
				{
					TargetVRF:              external.vrf,
					ImportVRFs:             []string{private.vrf},
					ImportPrefixes:         external.prefixes,
					ImportPrefixesNoExport: private.prefixes,
				},
			},
		},
		{
			name:  "firewall of a shared private network (shared/storage firewall)",
			input: "testdata/firewall_shared.yaml",
			want: []*importRule{
				{
					TargetVRF:      shared.vrf,
					ImportVRFs:     []string{inet.vrf},
					ImportPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes),
				},
				{
					TargetVRF:              inet.vrf,
					ImportVRFs:             []string{shared.vrf},
					ImportPrefixes:         concatPfxSlices(inet.prefixes),
					ImportPrefixesNoExport: shared.prefixes,
				},
				nil,
			},
		},
		{
			name:  "firewall of a private network with dmz network and internet (dmz firewall)",
			input: "testdata/firewall_dmz.yaml",
			want: []*importRule{
				{
					TargetVRF:      private.vrf,
					ImportVRFs:     []string{inet.vrf, dmz.vrf},
					ImportPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes, dmz.prefixes),
				},
				{
					TargetVRF:      dmz.vrf,
					ImportVRFs:     []string{private.vrf, inet.vrf},
					ImportPrefixes: concatPfxSlices(private.prefixes, dmz.prefixes, inet.destinations, inet.prefixes),
				},
				{
					TargetVRF:              inet.vrf,
					ImportVRFs:             []string{private.vrf, dmz.vrf},
					ImportPrefixes:         inet.prefixes,
					ImportPrefixesNoExport: concatPfxSlices(private.prefixes, dmz.prefixes),
				},
				nil,
			},
		},
		{
			name:  "firewall of a private network with dmz network (dmz app firewall)",
			input: "testdata/firewall_dmz_app.yaml",
			want: []*importRule{
				{
					TargetVRF:      private.vrf,
					ImportVRFs:     []string{dmz.vrf},
					ImportPrefixes: concatPfxSlices([]importPrefix{publicDefaultNet2}, dmz.prefixes, dmz.destinations),
				},
				{
					TargetVRF:      dmz.vrf,
					ImportVRFs:     []string{private.vrf},
					ImportPrefixes: concatPfxSlices(private.prefixes, dmz.prefixes),
				},
				nil,
			},
		},
		// {
		// 	name:  "firewall of a private network with dmz network and storage (dmz app firewall)",
		// 	input: "testdata/firewall_dmz_app_storage.yaml",
		// 	want: []*importRule{
		// 		{
		// 			targetVRF:      private.vrf,
		// 			importVRFs:     []string{dmz.vrf},
		// 			importPrefixes: concatPfxSlices([]importPrefix{publicDefaultNet2}, dmz.prefixes, dmz.destinations),
		// 		},
		// 		{
		// 			targetVRF:      dmz.vrf,
		// 			importVRFs:     []string{private.vrf},
		// 			importPrefixes: concatPfxSlices(private.prefixes, dmz.prefixes),
		// 		},
		// 		{
		// 			targetVRF:      shared.vrf,
		// 			importVRFs:     []string{private.vrf},
		// 			importPrefixes: concatPfxSlices(private.prefixes, shared.prefixes),
		// 		},
		// 		nil,
		// 	},
		// },
		{
			name:  "firewall with ipv6 private network and ipv6 internet network",
			input: "testdata/firewall_ipv6.yaml",
			want: []*importRule{
				{
					TargetVRF:      private6.vrf,
					ImportVRFs:     []string{inet.vrf, external.vrf, shared.vrf},
					ImportPrefixes: concatPfxSlices(inet6.destinations, external.destinations, []importPrefix{publicDefaultNetIPv6}, inet6.prefixes, external.prefixes, shared.prefixes),
				},
				{
					TargetVRF:      shared.vrf,
					ImportVRFs:     []string{private6.vrf},
					ImportPrefixes: concatPfxSlices(private6.prefixes, shared.prefixes),
				},
				{
					TargetVRF:              inet6.vrf,
					ImportVRFs:             []string{private.vrf},
					ImportPrefixes:         inet6.prefixes,
					ImportPrefixesNoExport: private6.prefixes,
				},
				nil,
				{
					TargetVRF:              external.vrf,
					ImportVRFs:             []string{private6.vrf},
					ImportPrefixes:         external.prefixes,
					ImportPrefixesNoExport: private6.prefixes,
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
			for i, network := range kb.Networks {
				got := importRulesForNetwork(kb, network)
				w := tt.want[i]
				if !reflect.DeepEqual(got, w) {
					if !reflect.DeepEqual(got.ImportVRFs, w.ImportVRFs) {
						t.Errorf("imported vrfs differ: %v", cmp.Diff(got.ImportVRFs, w.ImportVRFs))
					}
					if !reflect.DeepEqual(got.prefixLists(), w.prefixLists()) {
						t.Errorf("prefix lists differ: %v", cmp.Diff(got.prefixLists(), w.prefixLists()))
					}
					if !reflect.DeepEqual(got.routeMaps(), w.routeMaps()) {
						t.Errorf("route maps differ: %v", cmp.Diff(got.routeMaps(), w.routeMaps()))
					}
					t.Errorf("importRulesForNetwork() \ng: %v, \nw: %v", got, tt.want[i])
				}
			}
		})
	}
}
