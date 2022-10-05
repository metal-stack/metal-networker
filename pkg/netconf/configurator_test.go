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
		actual := NewConfigurator(test.kind, config{})
		assert.IsType(t, test.expected, actual)
	}
}
