package system

import (
	"testing"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf"
	"github.com/stretchr/testify/assert"
)

func TestChronyServiceEnabler_Enable(t *testing.T) {
	assert := assert.New(t)

	network := netconf.Network{Primary: false, Underlay: false, Destinationprefixes: []string{"0.0.0.0/0"}, Vrf: 104009}
	tests := []struct {
		kb              netconf.KnowledgeBase
		vrf             string
		isErrorExpected bool
	}{
		{kb: netconf.KnowledgeBase{Networks: []netconf.Network{network}},
			vrf:             "vrf104009",
			isErrorExpected: false},
		{kb: netconf.KnowledgeBase{Networks: []netconf.Network{}},
			vrf:             "",
			isErrorExpected: true},
	}

	for _, t := range tests {
		e, err := NewChronyServiceEnabler(t.kb)
		if t.isErrorExpected {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
		assert.Equal(t.vrf, e.VRF)
	}

}
