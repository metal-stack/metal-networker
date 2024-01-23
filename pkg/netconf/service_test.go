package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/metal-stack/metal-networker/pkg/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestServices(t *testing.T) {
	log := zaptest.NewLogger(t).Sugar()

	kb, err := New(log, "testdata/firewall.yaml")
	require.NoError(t, err)
	v := serviceValidator{}
	dsApplier, err := newDroptailerServiceApplier(*kb, v)
	require.NoError(t, err)
	fcApplier, err := newFirewallControllerServiceApplier(*kb, v)
	require.NoError(t, err)
	nodeExporterApplier, err := newNodeExporterServiceApplier(*kb, v)
	require.NoError(t, err)
	suApplier, err := newSuricataUpdateServiceApplier(*kb, v)
	require.NoError(t, err)
	nftablesExporterApplier, err := NewNftablesExporterServiceApplier(*kb, v)
	require.NoError(t, err)

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
		require.NoError(t, err)

		b := bytes.Buffer{}
		tpl := MustParseTpl(test.template)
		err = test.applier.Render(&b, *tpl)
		require.NoError(t, err)
		assert.Equal(t, string(expected), b.String())
	}
}
