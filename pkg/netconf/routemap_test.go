package netconf

import (
	"fmt"
	"reflect"
	"testing"

	"inet.af/netaddr"
)

type network struct {
	vrf          string
	prefixes     []importPrefix
	destinations []importPrefix
}

var (
	defaultRoute           = importPrefix{prefix: netaddr.MustParseIPPrefix("0.0.0.0/0"), policy: Permit}
	defaultRoute6          = importPrefix{prefix: netaddr.MustParseIPPrefix("::/0"), policy: Permit}
	externalNet            = importPrefix{prefix: netaddr.MustParseIPPrefix("100.127.129.0/24"), policy: Permit}
	externalDestinationNet = importPrefix{prefix: netaddr.MustParseIPPrefix("100.127.1.0/24"), policy: Permit}
	privateNet             = importPrefix{prefix: netaddr.MustParseIPPrefix("10.0.16.0/22"), policy: Permit}
	privateNet6            = importPrefix{prefix: netaddr.MustParseIPPrefix("2002::/64"), policy: Permit}
	sharedNet              = importPrefix{prefix: netaddr.MustParseIPPrefix("10.0.18.0/22"), policy: Permit}
	dmzNet                 = importPrefix{prefix: netaddr.MustParseIPPrefix("10.0.20.0/22"), policy: Permit}
	inetNet1               = importPrefix{prefix: netaddr.MustParseIPPrefix("185.1.2.0/24"), policy: Permit}
	inetNet2               = importPrefix{prefix: netaddr.MustParseIPPrefix("185.27.0.0/22"), policy: Permit}
	inetNet6               = importPrefix{prefix: netaddr.MustParseIPPrefix("2a02:c00:20::/45"), policy: Permit}
	publicDefaultNet       = importPrefix{prefix: netaddr.MustParseIPPrefix("185.1.2.3/32"), policy: Deny}
	publicDefaultNet2      = importPrefix{prefix: netaddr.MustParseIPPrefix("10.0.20.2/32"), policy: Deny}
	publicDefaultNetIPv6   = importPrefix{prefix: netaddr.MustParseIPPrefix("2a02:c00:20::1/128"), policy: Deny}

	private = network{
		vrf:      "vrf3981",
		prefixes: []importPrefix{privateNet},
	}

	private6 = network{
		vrf:      "vrf3981",
		prefixes: []importPrefix{privateNet6},
	}

	inet = network{
		vrf:          "vrf104009",
		prefixes:     []importPrefix{inetNet1, inetNet2},
		destinations: []importPrefix{defaultRoute},
	}

	inet6 = network{
		vrf:          "vrf104009",
		prefixes:     []importPrefix{inetNet6},
		destinations: []importPrefix{defaultRoute6},
	}

	external = network{
		vrf:          "vrf104010",
		destinations: []importPrefix{externalDestinationNet},
		prefixes:     []importPrefix{externalNet},
	}

	shared = network{
		vrf:      "vrf3982",
		prefixes: []importPrefix{sharedNet},
	}

	dmz = network{
		vrf:          "vrf3983",
		prefixes:     []importPrefix{dmzNet},
		destinations: []importPrefix{defaultRoute},
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
					targetVRF:      private.vrf,
					importVRFs:     []string{inet.vrf, external.vrf, shared.vrf},
					importPrefixes: concatPfxSlices(inet.destinations, external.destinations, []importPrefix{publicDefaultNet}, inet.prefixes, external.prefixes, shared.prefixes),
				},
				{
					targetVRF:      shared.vrf,
					importVRFs:     []string{private.vrf},
					importPrefixes: concatPfxSlices(private.prefixes, shared.prefixes),
				},
				{
					targetVRF:              inet.vrf,
					importVRFs:             []string{private.vrf},
					importPrefixes:         inet.prefixes,
					importPrefixesNoExport: private.prefixes,
				},
				nil,
				{
					targetVRF:              external.vrf,
					importVRFs:             []string{private.vrf},
					importPrefixes:         external.prefixes,
					importPrefixesNoExport: private.prefixes,
				},
			},
		},
		{
			name:  "firewall of a shared private network (shared/storage firewall)",
			input: "testdata/firewall_shared.yaml",
			want: []*importRule{
				{
					targetVRF:      shared.vrf,
					importVRFs:     []string{inet.vrf},
					importPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes),
				},
				{
					targetVRF:              inet.vrf,
					importVRFs:             []string{shared.vrf},
					importPrefixes:         concatPfxSlices(inet.prefixes),
					importPrefixesNoExport: shared.prefixes,
				},
				nil,
			},
		},
		{
			name:  "firewall of a private network with dmz network and internet (dmz firewall)",
			input: "testdata/firewall_dmz.yaml",
			want: []*importRule{
				{
					targetVRF:      private.vrf,
					importVRFs:     []string{inet.vrf, dmz.vrf},
					importPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes, dmz.prefixes),
				},
				{
					targetVRF:      dmz.vrf,
					importVRFs:     []string{private.vrf, inet.vrf},
					importPrefixes: concatPfxSlices(private.prefixes, dmz.prefixes, dmz.destinations, inet.prefixes),
				},
				{
					targetVRF:              inet.vrf,
					importVRFs:             []string{private.vrf, dmz.vrf},
					importPrefixes:         inet.prefixes,
					importPrefixesNoExport: concatPfxSlices(private.prefixes, dmz.prefixes),
				},
				nil,
			},
		},
		{
			name:  "firewall of a private network with dmz network (dmz app firewall)",
			input: "testdata/firewall_dmz_app.yaml",
			want: []*importRule{
				{
					targetVRF:      private.vrf,
					importVRFs:     []string{dmz.vrf},
					importPrefixes: concatPfxSlices([]importPrefix{publicDefaultNet2}, dmz.prefixes, dmz.destinations),
				},
				{
					targetVRF:      dmz.vrf,
					importVRFs:     []string{private.vrf},
					importPrefixes: concatPfxSlices(private.prefixes, dmz.prefixes),
				},
				nil,
			},
		},
		{
			name:  "firewall with ipv6 private network and ipv6 internet network",
			input: "testdata/firewall_ipv6.yaml",
			want: []*importRule{
				{
					targetVRF:      private6.vrf,
					importVRFs:     []string{inet.vrf, external.vrf, shared.vrf},
					importPrefixes: concatPfxSlices(inet6.destinations, external.destinations, []importPrefix{publicDefaultNetIPv6}, inet6.prefixes, external.prefixes, shared.prefixes),
				},
				{
					targetVRF:      shared.vrf,
					importVRFs:     []string{private6.vrf},
					importPrefixes: concatPfxSlices(private6.prefixes, shared.prefixes),
				},
				{
					targetVRF:              inet6.vrf,
					importVRFs:             []string{private.vrf},
					importPrefixes:         inet6.prefixes,
					importPrefixesNoExport: private6.prefixes,
				},
				nil,
				{
					targetVRF:              external.vrf,
					importVRFs:             []string{private6.vrf},
					importPrefixes:         external.prefixes,
					importPrefixesNoExport: private6.prefixes,
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
				if !reflect.DeepEqual(got, tt.want[i]) {
					fmt.Printf("import prefixes: g: %v\nw: %v\n", got.importPrefixes, tt.want[i].importPrefixes)
					fmt.Printf("no export: g: %v\nw: %v\n", got.importPrefixesNoExport, tt.want[i].importPrefixesNoExport)
					t.Errorf("importRulesForNetwork() got %v, wanted %v", got, tt.want[i])
				}
			}
		})
	}
}
