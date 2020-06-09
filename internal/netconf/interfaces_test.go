package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/metal-stack/metal-networker/pkg/net"
	"github.com/stretchr/testify/assert"
)

type FileRenderInfo struct {
	input            string
	expectedOutput   string
	configuratorType BareMetalType
	tpl              string
	newApplierFunc   func(BareMetalType, KnowledgeBase, string) net.Applier
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

		tpl := mustParseTpl(t.tpl)
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
			expectedOutput:   "testdata/lo.network.machine",
			configuratorType: Machine,
			tpl:              TplMachineIfaces,
			newApplierFunc:   NewIfacesConfigApplier,
		},
	}
	renderFilesAndVerifyExpectations(t, tests)
}
