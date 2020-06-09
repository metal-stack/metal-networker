package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/metal-stack/metal-networker/pkg/net"
	"github.com/stretchr/testify/assert"
)

func TestServices(t *testing.T) {
	assert := assert.New(t)

	kb := NewKnowledgeBase("testdata/firewall.yaml")
	v := ServiceValidator{}
	dsApplier, err := NewDroptailerServiceApplier(kb, v)
	assert.NoError(err)
	fpcApplier, err := NewFirewallPolicyControllerServiceApplier(kb, v)
	assert.NoError(err)
	nodeExporterApplier, err := NewNodeExporterServiceApplier(kb, v)
	assert.NoError(err)
	suApplier, err := NewSuricataUpdateServiceApplier(kb, v)
	assert.NoError(err)

	tests := []struct {
		applier  net.Applier
		expected string
		template string
	}{
		{
			applier:  dsApplier,
			expected: "testdata/droptailer.service",
			template: TplDroptailer,
		},
		{
			applier:  fpcApplier,
			expected: "testdata/firewall-policy-controller.service",
			template: TplFirewallPolicyController,
		},
		{
			applier:  nodeExporterApplier,
			expected: "testdata/node-exporter.service",
			template: TplNodeExporter,
		},
		{
			applier:  suApplier,
			expected: "testdata/suricata-update.service",
			template: TplSuricataUpdate,
		},
	}

	for _, test := range tests {
		expected, err := ioutil.ReadFile(test.expected)
		assert.NoError(err)

		b := bytes.Buffer{}
		tpl := mustParseTpl(test.template)
		err = test.applier.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}
