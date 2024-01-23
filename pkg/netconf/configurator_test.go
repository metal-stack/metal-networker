package netconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigurator(t *testing.T) {
	tests := []struct {
		kind     BareMetalType
		expected any
	}{
		{
			kind:     Firewall,
			expected: firewallConfigurator{},
		},
		{
			kind:     Machine,
			expected: machineConfigurator{},
		},
	}

	for _, tt := range tests {
		actual, err := NewConfigurator(tt.kind, config{}, false)
		require.NoError(t, err)
		assert.IsType(t, tt.expected, actual)
	}
}
