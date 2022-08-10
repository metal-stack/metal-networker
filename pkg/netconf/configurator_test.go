package netconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	for _, test := range tests {
		actual, err := NewConfigurator(test.kind, config{}, false)
		assert.NoError(t, err)
		assert.IsType(t, test.expected, actual)
	}
}
