package netconf

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileNftRules(t *testing.T) {

	tests := []struct {
		input          string
		expected       string
		enableDNSProxy bool
		forwardPolicy  ForwardPolicy
	}{
		{
			input:          "testdata/firewall.yaml",
			expected:       "testdata/nftrules",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyDrop,
		},
		{
			input:          "testdata/firewall.yaml",
			expected:       "testdata/nftrules_accept_forwarding",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyAccept,
		},
		{
			input:          "testdata/firewall_dmz.yaml",
			expected:       "testdata/nftrules_dmz",
			enableDNSProxy: true,
			forwardPolicy:  ForwardPolicyDrop,
		},
		{
			input:          "testdata/firewall_dmz_app.yaml",
			expected:       "testdata/nftrules_dmz_app",
			enableDNSProxy: true,
			forwardPolicy:  ForwardPolicyDrop,
		},
		{
			input:          "testdata/firewall_ipv6.yaml",
			expected:       "testdata/nftrules_ipv6",
			enableDNSProxy: true,
			forwardPolicy:  ForwardPolicyDrop,
		},
		{
			input:          "testdata/firewall_shared.yaml",
			expected:       "testdata/nftrules_shared",
			enableDNSProxy: true,
			forwardPolicy:  ForwardPolicyDrop,
		},
		{
			input:          "testdata/firewall_vpn.yaml",
			expected:       "testdata/nftrules_vpn",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyDrop,
		},
		{
			input:          "testdata/firewall_with_rules.yaml",
			expected:       "testdata/nftrules_with_rules",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyDrop,
		},
	}
	log := slog.Default()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			expected, err := os.ReadFile(tt.expected)
			require.NoError(t, err)

			kb, err := New(log, tt.input)
			require.NoError(t, err)

			a := newNftablesConfigApplier(*kb, nil, tt.enableDNSProxy, tt.forwardPolicy)
			b := bytes.Buffer{}

			tpl := MustParseTpl(TplNftables)
			err = a.Render(&b, *tpl)
			require.NoError(t, err)
			assert.Equal(t, string(expected), b.String())
		})
	}
}
