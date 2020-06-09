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
		expected string
		template string
	}{
		{expected: "testdata/nftrules.v4", template: TplNftablesV4},
		{expected: "testdata/nftrules.v6", template: TplNftablesV6},
	}

	for _, test := range tests {
		expected, err := ioutil.ReadFile(test.expected)
		assert.NoError(err)

		kb := NewKnowledgeBase("testdata/firewall.yaml")
		assert.NoError(err)

		a := NewNftablesConfigApplier(kb, nil)
		b := bytes.Buffer{}

		tpl := mustParseTpl(test.template)
		err = a.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}
