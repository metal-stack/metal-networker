package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"git.f-i-ts.de/cloud-native/metallib/network"

	"github.com/stretchr/testify/assert"
)

type FileRenderInfo struct {
	input            string
	expectedOutput   string
	configuratorType BareMetalType
	tpl              string
	newApplierFunc   func(BareMetalType, KnowledgeBase, string) network.Applier
}

func renderFilesAndVerifyExpectations(t *testing.T, tests []FileRenderInfo) {
	assert := assert.New(t)

	for _, t := range tests {
		expected, err := ioutil.ReadFile(t.expectedOutput)
		assert.NoError(err)

		kb := NewKnowledgeBase(t.input)
		assert.NoError(err)
		a := t.newApplierFunc(t.configuratorType, kb, "")
		b := bytes.Buffer{}

		s, err := ioutil.ReadFile(t.tpl)
		assert.NoError(err)
		tpl := template.Must(template.New(t.tpl).Parse(string(s)))
		err = a.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}

func TestCompileInterfaces(t *testing.T) {
	tests := []FileRenderInfo{
		{
			input:            "testdata/firewall.yaml",
			expectedOutput:   "testdata/interfaces.firewall",
			configuratorType: Firewall,
			tpl:              TplFirewallIfaces,
			newApplierFunc:   NewIfacesConfigApplier,
		},
		{
			input:            "testdata/machine.yaml",
			expectedOutput:   "testdata/interfaces.machine",
			configuratorType: Machine,
			tpl:              TplMachineIfaces,
			newApplierFunc:   NewIfacesConfigApplier,
		},
	}
	renderFilesAndVerifyExpectations(t, tests)
}
