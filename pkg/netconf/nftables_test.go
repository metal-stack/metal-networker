package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileNftRules(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "testdata/firewall.yaml",
			expected: "testdata/nftrules",
		},
		{
			input:    "testdata/firewall_dmz.yaml",
			expected: "testdata/nftrules_dmz",
		},
		{
			input:    "testdata/firewall_dmz_app.yaml",
			expected: "testdata/nftrules_dmz_app",
		},
		{
			input:    "testdata/firewall_ipv6.yaml",
			expected: "testdata/nftrules_ipv6",
		},
		{
			input:    "testdata/firewall_shared.yaml",
			expected: "testdata/nftrules_shared",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			expected, err := ioutil.ReadFile(tt.expected)
			assert.NoError(err)

			kb := NewKnowledgeBase(tt.input)
			assert.NoError(err)

			a := NewNftablesConfigApplier(kb, nil)
			b := bytes.Buffer{}

			tpl := mustParseTpl(TplNftables)
			err = a.Render(&b, *tpl)
			assert.NoError(err)
			assert.Equal(string(expected), b.String())
		})
	}
}
