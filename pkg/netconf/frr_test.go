package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestFrrConfigApplier(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedOutput   string
		configuratorType BareMetalType
		tpl              string
	}{
		{
			name:             "firewall of a shared private network",
			input:            "testdata/firewall_shared.yaml",
			expectedOutput:   "testdata/frr.conf.firewall_shared",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "standard firewall with private primary unshared network, private secondary shared network, internet and mpls",
			input:            "testdata/firewall.yaml",
			expectedOutput:   "testdata/frr.conf.firewall",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "dmz firewall with private primary unshared network, private secondary shared dmz network, internet and mpls",
			input:            "testdata/firewall_dmz.yaml",
			expectedOutput:   "testdata/frr.conf.firewall_dmz",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "dmz firewall with private primary unshared network, private secondary shared dmz network",
			input:            "testdata/firewall_dmz_app.yaml",
			expectedOutput:   "testdata/frr.conf.firewall_dmz_app",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "firewall with private primary unshared network, private secondary shared dmz network and private secondary shared storage network",
			input:            "testdata/firewall_dmz_app_storage.yaml",
			expectedOutput:   "testdata/frr.conf.firewall_dmz_app_storage",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "firewall with private primary unshared ipv6 network, private secondary shared ipv4 network, ipv6 internet and ipv4 mpls",
			input:            "testdata/firewall_ipv6.yaml",
			expectedOutput:   "testdata/frr.conf.firewall_ipv6",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			name:             "standard machine",
			input:            "testdata/machine.yaml",
			expectedOutput:   "testdata/frr.conf.machine",
			configuratorType: Machine,
			tpl:              TplMachineFRR,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			log := zaptest.NewLogger(t).Sugar()
			kb, err := New(log, test.input)
			assert.NoError(t, err)
			a := NewFrrConfigApplier(test.configuratorType, *kb, "")
			b := bytes.Buffer{}

			tpl := mustParseTpl(test.tpl)
			err = a.Render(&b, *tpl)
			assert.NoError(t, err)

			// eases adjustment of test fixtures
			// just remove old test fixture after a code change
			// let the new fixtures get generated
			// check them manually before commit
			if _, err := os.Stat(test.expectedOutput); os.IsNotExist(err) {
				err = os.WriteFile(test.expectedOutput, b.Bytes(), fileModeDefault)
				assert.NoError(t, err)
				return
			}

			expected, err := os.ReadFile(test.expectedOutput)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), b.String())
		})
	}
}

func TestFRRValidator_Validate(t *testing.T) {
	assert := assert.New(t)
	log := zaptest.NewLogger(t).Sugar()

	validator := FRRValidator{
		log: log,
	}
	actual := validator.Validate()
	assert.NotNil(actual)
	assert.NotNil(actual.Error())
}
