package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

		kb := NewKnowledgeBase("testdata/firewall.yaml", zap.NewNop().Sugar())
		assert.NoError(err)

		a := NewNftablesConfigApplier(kb, nil)
		b := bytes.Buffer{}

		f := test.template
		s, err := ioutil.ReadFile(f)
		assert.NoError(err)
		tpl := template.Must(template.New(f).Parse(string(s)))
		err = a.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}
