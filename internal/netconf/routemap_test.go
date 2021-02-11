package netconf

import (
	"reflect"
	"testing"

	"inet.af/netaddr"
)

type network struct {
	vrf          string
	prefixes     []netaddr.IPPrefix
	destinations []netaddr.IPPrefix
}

var (
	defaultRoute           = netaddr.MustParseIPPrefix("0.0.0.0/0")
	externalNet            = netaddr.MustParseIPPrefix("100.127.129.0/24")
	externalDestinationNet = netaddr.MustParseIPPrefix("100.127.1.0/24")
	privateNet             = netaddr.MustParseIPPrefix("10.0.16.0/22")
	sharedNet              = netaddr.MustParseIPPrefix("10.0.18.0/22")
	inetNet1               = netaddr.MustParseIPPrefix("185.1.2.0/24")
	inetNet2               = netaddr.MustParseIPPrefix("185.27.0.0/22")

	private = network{
		vrf:      "vrf3981",
		prefixes: []netaddr.IPPrefix{privateNet},
	}

	inet = network{
		vrf:          "vrf104009",
		prefixes:     []netaddr.IPPrefix{inetNet1, inetNet2},
		destinations: []netaddr.IPPrefix{defaultRoute},
	}

	external = network{
		vrf:          "vrf104010",
		destinations: []netaddr.IPPrefix{externalDestinationNet},
		prefixes:     []netaddr.IPPrefix{externalNet},
	}

	shared = network{
		vrf:      "vrf3982",
		prefixes: []netaddr.IPPrefix{sharedNet},
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
					importPrefixes: concatPfxSlices(inet.destinations, external.destinations, shared.prefixes),
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
			name:  "firewall of a shared private network",
			input: "testdata/firewall_shared.yaml",
			want: []*importRule{
				{
					targetVRF:      shared.vrf,
					importVRFs:     []string{inet.vrf},
					importPrefixes: concatPfxSlices(inet.destinations),
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
		// {
		// 	name:  "firewall with private primary unshared ipv6 network, private secondary shared ipv4 network, ipv6 internet and ipv4 mpls",
		// 	input: "testdata/firewall_ipv6.yaml",
		// 	want: []*Import{
		// TODO
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := NewKnowledgeBase(tt.input)
			for i, network := range kb.Networks {
				got := importRulesForNetwork(kb, network)
				if !reflect.DeepEqual(got, tt.want[i]) {
					t.Errorf("importRulesForNetwork() got %v, wanted %v", got, tt.want[i])
				}
			}
		})
	}
}
