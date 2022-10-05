package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileNftRules(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input          string
		expected       string
		enableDNSProxy bool
	}{
		{
			input:          "testdata/firewall.yaml",
			expected:       "testdata/nftrules",
			enableDNSProxy: false,
		},
		{
			input:          "testdata/firewall_dmz.yaml",
			expected:       "testdata/nftrules_dmz",
			enableDNSProxy: true,
		},
		{
			input:          "testdata/firewall_dmz_app.yaml",
			expected:       "testdata/nftrules_dmz_app",
			enableDNSProxy: true,
		},
		{
			input:          "testdata/firewall_ipv6.yaml",
			expected:       "testdata/nftrules_ipv6",
			enableDNSProxy: true,
		},
		{
			input:          "testdata/firewall_shared.yaml",
			expected:       "testdata/nftrules_shared",
			enableDNSProxy: true,
		},
		{
			input:          "testdata/firewall_vpn.yaml",
			expected:       "testdata/nftrules_vpn",
			enableDNSProxy: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			expected, err := os.ReadFile(tt.expected)
			assert.NoError(err)

			kb, err := NewKnowledgeBase(tt.input)
			assert.NoError(err)

			a := NewNftablesConfigApplier(*kb, nil, tt.enableDNSProxy)
			b := bytes.Buffer{}

			tpl := mustParseTpl(TplNftables)
			err = a.Render(&b, *tpl)
			assert.NoError(err)
			assert.Equal(string(expected), b.String())
		})
	}
}
