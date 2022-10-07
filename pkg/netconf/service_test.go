package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/metal-stack/metal-networker/pkg/net"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestServices(t *testing.T) {
	assert := assert.New(t)
	log := zaptest.NewLogger(t).Sugar()

	kb, err := New(log, "testdata/firewall.yaml")
	assert.NoError(err)
	v := serviceValidator{}
	dsApplier, err := newDroptailerServiceApplier(*kb, v)
	assert.NoError(err)
	fcApplier, err := newFirewallControllerServiceApplier(*kb, v)
	assert.NoError(err)
	nodeExporterApplier, err := newNodeExporterServiceApplier(*kb, v)
	assert.NoError(err)
	suApplier, err := newSuricataUpdateServiceApplier(*kb, v)
	assert.NoError(err)
	nftablesExporterApplier, err := NewNftablesExporterServiceApplier(*kb, v)
	assert.NoError(err)

	tests := []struct {
		applier  net.Applier
		expected string
		template string
	}{
		{
			applier:  dsApplier,
			expected: "testdata/droptailer.service",
			template: tplDroptailer,
		},
		{
			applier:  fcApplier,
			expected: "testdata/firewall-controller.service",
			template: tplFirewallController,
		},
		{
			applier:  nodeExporterApplier,
			expected: "testdata/node-exporter.service",
			template: tplNodeExporter,
		},
		{
			applier:  nftablesExporterApplier,
			expected: "testdata/nftables-exporter.service",
			template: tplNftablesExporter,
		},
		{
			applier:  suApplier,
			expected: "testdata/suricata-update.service",
			template: tplSuricataUpdate,
		},
	}

	for _, test := range tests {
		expected, err := os.ReadFile(test.expected)
		assert.NoError(err)

		b := bytes.Buffer{}
		tpl := mustParseTpl(test.template)
		err = test.applier.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}
