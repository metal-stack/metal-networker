package netconf

import (
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-hammer/pkg/api"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChronyServiceEnabler_Enable(t *testing.T) {
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

	for _, tt := range tests {
		e, err := newChronyServiceEnabler(tt.kb)
		if tt.isErrorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		assert.Equal(t, tt.vrf, e.vrf)
	}
}
