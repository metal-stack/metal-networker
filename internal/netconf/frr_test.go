package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name:             "standard machine",
			input:            "testdata/machine.yaml",
			expectedOutput:   "testdata/frr.conf.machine",
			configuratorType: Machine,
			tpl:              TplMachineFRR,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected, err := ioutil.ReadFile(test.expectedOutput)
			assert.NoError(t, err)

			kb := NewKnowledgeBase(test.input)
			assert.NoError(t, err)
			a := NewFrrConfigApplier(test.configuratorType, kb, "")
			b := bytes.Buffer{}

			tpl := mustParseTpl(test.tpl)
			err = a.Render(&b, *tpl)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), b.String())
		})
	}
}

func TestFRRValidator_Validate(t *testing.T) {
	assert := assert.New(t)

	validator := FRRValidator{}
	actual := validator.Validate()
	assert.NotNil(actual)
	assert.NotNil(actual.Error())
}
