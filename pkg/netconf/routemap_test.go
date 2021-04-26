package netconf

import (
	"fmt"
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
	defaultRoute6          = netaddr.MustParseIPPrefix("::/0")
	externalNet            = netaddr.MustParseIPPrefix("100.127.129.0/24")
	externalDestinationNet = netaddr.MustParseIPPrefix("100.127.1.0/24")
	privateNet             = netaddr.MustParseIPPrefix("10.0.16.0/22")
	privateNet6            = netaddr.MustParseIPPrefix("2002::/64")
	sharedNet              = netaddr.MustParseIPPrefix("10.0.18.0/22")
	dmzNet                 = netaddr.MustParseIPPrefix("10.0.20.0/22")
	inetNet1               = netaddr.MustParseIPPrefix("185.1.2.0/24")
	inetNet2               = netaddr.MustParseIPPrefix("185.27.0.0/22")
	inetNet6               = netaddr.MustParseIPPrefix("2a02:c00:20::/45")

	private = network{
		vrf:      "vrf3981",
		prefixes: []netaddr.IPPrefix{privateNet},
	}

	private6 = network{
		vrf:      "vrf3981",
		prefixes: []netaddr.IPPrefix{privateNet6},
	}

	inet = network{
		vrf:          "vrf104009",
		prefixes:     []netaddr.IPPrefix{inetNet1, inetNet2},
		destinations: []netaddr.IPPrefix{defaultRoute},
	}

	inet6 = network{
		vrf:          "vrf104009",
		prefixes:     []netaddr.IPPrefix{inetNet6},
		destinations: []netaddr.IPPrefix{defaultRoute6},
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

	dmz = network{
		vrf:          "vrf3983",
		prefixes:     []netaddr.IPPrefix{dmzNet},
		destinations: []netaddr.IPPrefix{defaultRoute},
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
			name:  "firewall of a shared private network (shared/storage firewall)",
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
		{
			name:  "firewall of a private network with dmz network and internet (dmz firewall)",
			input: "testdata/firewall_dmz.yaml",
			want: []*importRule{
				{
					targetVRF:      private.vrf,
					importVRFs:     []string{inet.vrf, dmz.vrf},
					importPrefixes: concatPfxSlices(inet.destinations, dmz.prefixes),
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
					importPrefixes: concatPfxSlices(dmz.prefixes, dmz.destinations),
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
					importPrefixes: concatPfxSlices(inet6.destinations, external.destinations, shared.prefixes),
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
					fmt.Printf("g: %v\nw: %v\n", got.importPrefixes, tt.want[i].importPrefixes)
					fmt.Printf("g: %v\nw: %v\n", got.importPrefixesNoExport, tt.want[i].importPrefixesNoExport)
					t.Errorf("importRulesForNetwork() got %v, wanted %v", got, tt.want[i])
				}
			}
		})
	}
}
