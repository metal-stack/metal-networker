package netconf

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrrConfigApplier(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		frrVersion       *semver.Version
		expectedOutput   string
		configuratorType BareMetalType
		tpl              string
	}{
		{
			name:             "firewall of a shared private network",
			input:            "testdata/firewall_shared.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.firewall_shared",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "standard firewall with private primary unshared network, private secondary shared network, internet and mpls",
			input:            "testdata/firewall.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.firewall",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "dmz firewall with private primary unshared network, private secondary shared dmz network, internet and mpls",
			input:            "testdata/firewall_dmz.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.firewall_dmz",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "dmz firewall with private primary unshared network, private secondary shared dmz network",
			input:            "testdata/firewall_dmz_app.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.firewall_dmz_app",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "firewall with private primary unshared network, private secondary shared dmz network and private secondary shared storage network",
			input:            "testdata/firewall_dmz_app_storage.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.firewall_dmz_app_storage",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "firewall with private primary unshared ipv6 network, private secondary shared ipv4 network, ipv6 internet and ipv4 mpls",
			input:            "testdata/firewall_ipv6.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.firewall_ipv6",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "firewall with private primary unshared ipv6 network, private secondary shared ipv4 network, dualstack internet and ipv4 mpls",
			input:            "testdata/firewall_dualstack.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.firewall_dualstack",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "standard machine",
			input:            "testdata/machine.yaml",
			frrVersion:       semver.MustParse("8.5"),
			expectedOutput:   "testdata/frr.conf.machine",
			configuratorType: Machine,
			tpl:              TplMachineFRR,
		},
		{
			name:             "standard firewall with lower frr version",
			input:            "testdata/firewall.yaml",
			frrVersion:       semver.MustParse("9.0.5-0"),
			expectedOutput:   "testdata/frr.conf.firewall_frr-9",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "standard firewall with higher frr version",
			input:            "testdata/firewall.yaml",
			frrVersion:       semver.MustParse("10.1.5"),
			expectedOutput:   "testdata/frr.conf.firewall_frr-10",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			log := slog.Default()
			kb, err := New(log, test.input)
			require.NoError(t, err)
			a := NewFrrConfigApplier(test.configuratorType, *kb, "", test.frrVersion)
			b := bytes.Buffer{}

			tpl := MustParseTpl(test.tpl)
			err = a.Render(&b, *tpl)
			require.NoError(t, err)

			// eases adjustment of test fixtures
			// just remove old test fixture after a code change
			// let the new fixtures get generated
			// check them manually before commit
			if _, err := os.Stat(test.expectedOutput); os.IsNotExist(err) {
				err = os.WriteFile(test.expectedOutput, b.Bytes(), fileModeDefault)
				require.NoError(t, err)
				return
			}

			expected, err := os.ReadFile(test.expectedOutput)
			require.NoError(t, err)
			assert.Equal(t, string(expected), b.String())
		})
	}
}

func TestFRRValidator_Validate(t *testing.T) {
	validator := frrValidator{
		log: slog.Default(),
	}
	actual := validator.Validate()
	require.Error(t, actual)
}
