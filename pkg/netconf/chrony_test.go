package netconf

import (
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-hammer/pkg/api"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"github.com/stretchr/testify/assert"
)

func TestChronyServiceEnabler_Enable(t *testing.T) {
	assert := assert.New(t)

	vrf := int64(104009)
	external := mn.External
	network := &models.V1MachineNetwork{Networktype: &external, Destinationprefixes: []string{IPv4ZeroCIDR}, Vrf: &vrf}
	tests := []struct {
		kb              config
		vrf             string
		isErrorExpected bool
	}{
		{
			kb:              config{InstallerConfig: api.InstallerConfig{Networks: []*models.V1MachineNetwork{network}}},
			vrf:             "vrf104009",
			isErrorExpected: false,
		},
		{
			kb:              config{InstallerConfig: api.InstallerConfig{Networks: []*models.V1MachineNetwork{}}},
			vrf:             "",
			isErrorExpected: true,
		},
	}

	for _, t := range tests {
		e, err := newChronyServiceEnabler(t.kb)
		if t.isErrorExpected {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
		assert.Equal(t.vrf, e.vrf)
	}
}
