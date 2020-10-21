package netconf

import (
	"fmt"
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
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
	assert.True(*n.Private)
	assert.Equal(mn.PrivatePrimaryUnshared, *n.Networktype)
	assert.Equal(int64(3981), *n.Vrf)

	// private shared network
	n = d.Networks[1]
	assert.Equal(1, len(n.Ips))
	assert.Equal("10.0.18.2", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("10.0.18.0/22", n.Prefixes[0])
	assert.True(*n.Private)
	assert.Equal(mn.PrivateSecondaryShared, *n.Networktype)
	assert.Equal(int64(3982), *n.Vrf)

	// public networks
	n = d.Networks[2]
	assert.Equal(1, len(n.Destinationprefixes))
	assert.Equal(AllZerosCIDR, n.Destinationprefixes[0])
	assert.Equal(1, len(n.Ips))
	assert.Equal("185.1.2.3", n.Ips[0])
	assert.Equal(2, len(n.Prefixes))
	assert.Equal("185.1.2.0/24", n.Prefixes[0])
	assert.Equal("185.27.0.0/22", n.Prefixes[1])
	assert.False(*n.Underlay)
	assert.False(*n.Private)
	assert.True(*n.Nat)
	assert.Equal(mn.External, *n.Networktype)
	assert.Equal(int64(104009), *n.Vrf)

	// underlay network
	n = d.Networks[3]
	assert.Equal(int64(4200003073), *n.Asn)
	assert.Equal(1, len(n.Ips))
	assert.Equal("10.1.0.1", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("10.0.12.0/22", n.Prefixes[0])
	assert.True(*n.Underlay)
	assert.Equal(mn.Underlay, *n.Networktype)

	// public network mpls
	n = d.Networks[4]
	assert.Equal(1, len(n.Destinationprefixes))
	assert.Equal("100.127.1.0/24", n.Destinationprefixes[0])
	assert.Equal(1, len(n.Ips))
	assert.Equal("100.127.129.1", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("100.127.129.0/24", n.Prefixes[0])
	assert.False(*n.Underlay)
	assert.False(*n.Private)
	assert.True(*n.Nat)
	assert.Equal(mn.External, *n.Networktype)
	assert.Equal(int64(104010), *n.Vrf)
}

var (
	boolTrue  = true
	boolFalse = false
	asn0      = int64(0)
	asn1      = int64(1011209)
	vrf0      = int64(0)
	vrf1      = int64(1011209)
)

func stubKnowledgeBase() KnowledgeBase {
	privateNetID := "private"
	underlayNetID := "underlay"
	privatePrimaryUnshared := mn.PrivatePrimaryUnshared
	underlay := mn.Underlay
	external := mn.External
	return KnowledgeBase{Networks: []models.V1MachineNetwork{
		{Private: &boolTrue, Networktype: &privatePrimaryUnshared, Ips: []string{"10.0.0.1"}, Asn: &asn1, Vrf: &vrf1, Networkid: &privateNetID},
		{Underlay: &boolTrue, Networktype: &underlay, Ips: []string{"10.0.0.1"}, Asn: &asn1, Vrf: &vrf0, Networkid: &underlayNetID},
		{Private: &boolFalse, Networktype: &external, Underlay: &boolFalse, Destinationprefixes: []string{"10.0.0.1/24"}, Asn: &asn1, Vrf: &vrf1, Networkid: &underlayNetID},
	}, Nics: []NIC{{Mac: "00:00:00:00:00:00"}}}
}

func TestKnowledgeBase_Validate(t *testing.T) {
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
			kb:    maskPrivatePrimaryNetworks(stubKnowledgeBase()),
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

	for i, test := range tests {
		for _, kind := range test.kinds {
			t.Run(fmt.Sprintf("testcase %d - kind %v", i, kind), func(t *testing.T) {
				actualErr := test.kb.Validate(kind)
				if test.expectedErrMsg == "" {
					assert.NoError(t, actualErr)
					return
				}
				assert.EqualError(t, actualErr, test.expectedErrMsg, "expected error: %s", test.expectedErrMsg)
			})
		}
	}
}

func stripVRFValueOfNonUnderlayNetworks(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		// underlay runs in default vrf and no name is required
		if kb.Networks[i].Underlay != nil && *kb.Networks[i].Underlay {
			continue
		}
		vrf := int64(0)
		kb.Networks[i].Vrf = &vrf
	}
	return kb
}

// It makes no sense to have an public network without destination prefixes.
// Destination prefixes are used to import routes from the public network.
// Without route import there is no communication into that public network.
func stripDestinationPrefixesFromPublicNetworks(kb KnowledgeBase) KnowledgeBase {
	kb.Networks[0].Nat = &boolTrue
	for i := 0; i < len(kb.Networks); i++ {
		if kb.Networks[i].Underlay != nil && !*kb.Networks[i].Underlay && kb.Networks[i].Private != nil && !*kb.Networks[i].Private {
			kb.Networks[i].Destinationprefixes = []string{}
		}
	}
	return kb
}

func setupIllegalNat(kb KnowledgeBase) KnowledgeBase {
	kb.Networks[0].Nat = &boolTrue
	for i := 0; i < len(kb.Networks); i++ {
		if kb.Networks[i].Private != nil && *kb.Networks[i].Private {
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
		if kb.Networks[i].Underlay != nil && *kb.Networks[i].Underlay {
			kb.Networks[i].Asn = &asn0
		}
	}
	return kb
}

func stripPrivateNetworkASN(kb KnowledgeBase) KnowledgeBase {
	for i := 0; i < len(kb.Networks); i++ {
		if kb.Networks[i].Private != nil && *kb.Networks[i].Private {
			kb.Networks[i].Asn = &asn0
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
	kb.Networks = []models.V1MachineNetwork{}
	return kb
}

func maskUnderlayNetworks(kb KnowledgeBase) KnowledgeBase {
	privateSecondary := mn.PrivateSecondaryShared
	for i, n := range kb.Networks {
		if n.Networktype != nil && *n.Networktype == mn.Underlay {
			kb.Networks[i].Underlay = &boolFalse
			kb.Networks[i].Networktype = &privateSecondary
			// avoid to run into validation error for absent vrf
			kb.Networks[i].Vrf = &vrf1
		}
	}
	return kb
}

func maskPrivatePrimaryNetworks(kb KnowledgeBase) KnowledgeBase {
	privateUnshared := mn.PrivatePrimaryUnshared
	for i := range kb.Networks {
		kb.Networks[i].Networktype = &privateUnshared
	}
	return kb
}
