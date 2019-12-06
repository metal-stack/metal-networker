package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestFirewallPolicyController(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		expected string
		template string
	}{
		{expected: "testdata/firewall-policy-controller.service", template: TplFirewallPolicyControllerService},
	}

	for _, test := range tests {
		expected, err := ioutil.ReadFile(test.expected)
		assert.NoError(err)

		kb := NewKnowledgeBase("testdata/firewall.yaml")
		assert.NoError(err)

		v := ServiceValidator{}
		a, _ := NewFirewallPolicyControllerServiceApplier(kb, v)
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
