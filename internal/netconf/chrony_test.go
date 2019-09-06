package netconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChronyServiceEnabler_Enable(t *testing.T) {
	assert := assert.New(t)

	network := Network{Private: false, Underlay: false, Destinationprefixes: []string{AllZerosCIDR}, Vrf: 104009}
	tests := []struct {
		kb              KnowledgeBase
		vrf             string
		isErrorExpected bool
	}{
		{kb: KnowledgeBase{Networks: []Network{network}},
			vrf:             "vrf104009",
			isErrorExpected: false},
		{kb: KnowledgeBase{Networks: []Network{}},
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
