package netconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigurator(t *testing.T) {
	tests := []struct {
		kind     BareMetalType
		expected interface{}
	}{
		{
			kind:     Firewall,
			expected: FirewallConfigurator{},
		},
		{
			kind:     Machine,
			expected: MachineConfigurator{},
		},
	}

	for _, test := range tests {
		actual := NewConfigurator(test.kind, KnowledgeBase{})
		assert.IsType(t, test.expected, actual)
	}
}
