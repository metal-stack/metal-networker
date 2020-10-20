package netconf

import (
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestChronyServiceEnabler_Enable(t *testing.T) {
	assert := assert.New(t)

	vrf := int64(104009)
	network := models.V1MachineNetwork{Networktype: &External, Destinationprefixes: []string{AllZerosCIDR}, Vrf: &vrf}
	tests := []struct {
		kb              KnowledgeBase
		vrf             string
		isErrorExpected bool
	}{
		{kb: KnowledgeBase{Networks: []models.V1MachineNetwork{network}},
			vrf:             "vrf104009",
			isErrorExpected: false},
		{kb: KnowledgeBase{Networks: []models.V1MachineNetwork{}},
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
