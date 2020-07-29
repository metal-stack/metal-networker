package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrrConfigApplier(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input            string
		expectedOutput   string
		configuratorType BareMetalType
		tpl              string
	}{
		{
			input:            "testdata/firewall.yaml",
			expectedOutput:   "testdata/frr.conf.firewall",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
		},
		{
			input:            "testdata/machine.yaml",
			expectedOutput:   "testdata/frr.conf.machine",
			configuratorType: Machine,
			tpl:              TplMachineFRR,
		},
	}
	for _, t := range tests {
		expected, err := ioutil.ReadFile(t.expectedOutput)
		assert.NoError(err)

		kb := NewKnowledgeBase(t.input)
		assert.NoError(err)
		a := NewFrrConfigApplier(t.configuratorType, kb, "")
		b := bytes.Buffer{}

		tpl := mustParseTpl(t.tpl)
		err = a.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}

func TestFRRValidator_Validate(t *testing.T) {
	assert := assert.New(t)

	validator := FRRValidator{}
	actual := validator.Validate()
	assert.NotNil(actual)
	assert.NotNil(actual.Error())
}
