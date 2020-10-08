package netconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustNewKnowledgeBase(t *testing.T) KnowledgeBase {
	assert := assert.New(t)

	d := NewKnowledgeBase("testdata/firewall.yaml")
	assert.NotNil(d)

	return d
}

func TestNewKnowledgeBase(t *testing.T) {
	assert := assert.New(t)

	d := mustNewKnowledgeBase(t)

	assert.Equal("firewall", d.Hostname)
	assert.NotEmpty(d.Networks)
	assert.Equal(5, len(d.Networks))

	// private network
	n := d.Networks[0]
	assert.Equal(1, len(n.Ips))
	assert.Equal("10.0.16.2", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("10.0.16.0/22", n.Prefixes[0])
	assert.True(n.Private)
	assert.False(n.Shared)
	assert.Equal(3981, n.Vrf)

	// private shared network
	n = d.Networks[1]
	assert.Equal(1, len(n.Ips))
	assert.Equal("10.0.18.2", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("10.0.18.0/22", n.Prefixes[0])
	assert.True(n.Private)
	assert.True(n.Shared)
	assert.Equal(3982, n.Vrf)

	// public networks
	n = d.Networks[2]
	assert.Equal(1, len(n.Destinationprefixes))
	assert.Equal(AllZerosCIDR, n.Destinationprefixes[0])
	assert.Equal(1, len(n.Ips))
	assert.Equal("185.1.2.3", n.Ips[0])
	assert.Equal(2, len(n.Prefixes))
	assert.Equal("185.1.2.0/24", n.Prefixes[0])
	assert.Equal("185.27.0.0/22", n.Prefixes[1])
	assert.False(n.Underlay)
	assert.False(n.Private)
	assert.True(n.Nat)
	assert.Equal(104009, n.Vrf)

	// underlay network
	n = d.Networks[3]
	assert.Equal(int64(4200003073), n.Asn)
	assert.Equal(1, len(n.Ips))
	assert.Equal("10.1.0.1", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("10.0.12.0/22", n.Prefixes[0])
	assert.True(n.Underlay)

	// public network mpls
	n = d.Networks[4]
	assert.Equal(1, len(n.Destinationprefixes))
	assert.Equal("100.127.1.0/24", n.Destinationprefixes[0])
	assert.Equal(1, len(n.Ips))
	assert.Equal("100.127.129.1", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("100.127.129.0/24", n.Prefixes[0])
	assert.False(n.Underlay)
	assert.False(n.Private)
	assert.True(n.Nat)
	assert.Equal(104010, n.Vrf)
}

func stubKnowledgeBase() KnowledgeBase {
	return KnowledgeBase{Networks: []Network{
		{Private: true, Ips: []string{"10.0.0.1"}, Asn: 1011209, Vrf: 1011209},
		{Underlay: true, Ips: []string{"10.0.0.1"}, Asn: 1011209, Vrf: 0},
		{Private: false, Underlay: false, Destinationprefixes: []string{"10.0.0.1/24"}, Asn: 1011209, Vrf: 1011209},
	}, Nics: []NIC{{Mac: "00:00:00:00:00:00"}}}
}

func TestKnowledgeBase_Validate(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		expectedErrMsg string
		kb             KnowledgeBase
		kinds          []BareMetalType
	}{{
		expectedErrMsg: "",
		kb:             stubKnowledgeBase(),
		kinds:          []BareMetalType{Firewall, Machine},
	},
		{
			expectedErrMsg: "expectation at least one network is present failed",
			kb:             stripNetworks(stubKnowledgeBase()),
			kinds:          []BareMetalType{Firewall, Machine},
		},
		{
			expectedErrMsg: "at least one IP must be present to be considered as LOOPBACK IP (" +
				"'private: true' network IP for machine, 'underlay: true' network IP for firewall",
			kb:    stripIPs(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall, Machine},
		},
		{expectedErrMsg: "expectation exactly one underlay network is present failed",
			kb:    maskUnderlayNetworks(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall}},
		{expectedErrMsg: "expectation exactly one 'private: true' network is present failed",
			kb:    maskPrivateNetworks(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall, Machine}},
		{expectedErrMsg: "'asn' of private (machine) resp. underlay (firewall) network must not be missing",
			kb:    stripPrivateNetworkASN(stubKnowledgeBase()),
			kinds: []BareMetalType{Machine}},
		{expectedErrMsg: "'asn' of private (machine) resp. underlay (firewall) network must not be missing",
			kb:    stripUnderlayNetworkASN(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall}},
		{expectedErrMsg: "at least one 'nics/nic' definition must be present",
			kb:    stripNICs(stubKnowledgeBase()),
			kinds: []BareMetalType{Machine}},
		{expectedErrMsg: "each 'nic' definition must contain a valid 'mac'",
			kb:    stripMACs(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall, Machine}},
		{expectedErrMsg: "private network must not lack prefixes since nat is required",
			kb:    setupIllegalNat(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall}},
		{expectedErrMsg: "non-private, non-underlay networks must contain destination prefix(es) to make any sense of it",
			kb:    stripDestinationPrefixesFromPublicNetworks(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall}},
		{expectedErrMsg: "networks with 'underlay: false' must contain a value vor 'vrf' as it is used for BGP",
			kb:    stripVRFValueOfNonUnderlayNetworks(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall}},
		{expectedErrMsg: "each 'nic' definition must contain a valid 'mac'",
			kb:    unlegalizeMACs(stubKnowledgeBase()),
			kinds: []BareMetalType{Firewall, Machine}},
	}

	for _, test := range tests {
		for _, kind := range test.kinds {
			actualErr := test.kb.Validate(kind)
			if test.expectedErrMsg == "" {
				assert.NoError(actualErr)
				continue
			}
			assert.EqualError(actualErr, test.expectedErrMsg, "expected error: %s", test.expectedErrMsg)
		}
	}
}

func stripVRFValueOfNonUnderlayNetworks(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		// underlay runs in default vrf and no name is required
		if kb.Networks[i].Underlay {
			continue
		}
		kb.Networks[i].Vrf = 0
	}
	return kb
}

// It makes no sense to have an public network without destination prefixes.
// Destination prefixes are used to import routes from the public network.
// Without route import there is no communication into that public network.
func stripDestinationPrefixesFromPublicNetworks(kb KnowledgeBase) KnowledgeBase {
	kb.Networks[0].Nat = true
	for i := 0; i < len(kb.Networks); i++ {
		if !kb.Networks[i].Underlay && !kb.Networks[i].Private {
			kb.Networks[i].Destinationprefixes = []string{}
		}
	}
	return kb
}

func setupIllegalNat(kb KnowledgeBase) KnowledgeBase {
	kb.Networks[0].Nat = true
	for i := 0; i < len(kb.Networks); i++ {
		if kb.Networks[i].Private {
			kb.Networks[i].Prefixes = []string{}
		}
	}
	return kb
}

func unlegalizeMACs(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Nics); i++ {
		kb.Nics[i].Mac = "1:2.3"
	}
	return kb
}

func stripMACs(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Nics); i++ {
		kb.Nics[i].Mac = ""
	}
	return kb
}

func stripNICs(kb KnowledgeBase) KnowledgeBase {
	kb.Nics = []NIC{}
	return kb
}

func stripUnderlayNetworkASN(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		if kb.Networks[i].Underlay {
			kb.Networks[i].Asn = 0
		}
	}
	return kb
}

func stripPrivateNetworkASN(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		if kb.Networks[i].Private {
			kb.Networks[i].Asn = 0
		}
	}
	return kb
}

func stripIPs(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		kb.Networks[i].Ips = []string{}
	}
	return kb
}

func stripNetworks(kb KnowledgeBase) KnowledgeBase {
	kb.Networks = []Network{}
	return kb
}

func maskUnderlayNetworks(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		kb.Networks[i].Underlay = false
		// avoid to run into validation error for absent vrf
		kb.Networks[i].Vrf = 10112009
	}
	return kb
}

func maskPrivateNetworks(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		kb.Networks[i].Private = false
	}
	return kb
}
