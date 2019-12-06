package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"git.f-i-ts.de/cloud-native/metallib/network"
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

	tests := []struct {
		applier  network.Applier
		expected string
		template string
	}{
		{
			applier:  dsApplier,
			expected: "testdata/droptailer.service",
			template: TplDroptailerService,
		},
		{
			applier:  fpcApplier,
			expected: "testdata/firewall-policy-controller.service",
			template: TplFirewallPolicyControllerService,
		},
	}

	for _, test := range tests {
		expected, err := ioutil.ReadFile(test.expected)
		assert.NoError(err)

		b := bytes.Buffer{}
		f := test.template
		s, err := ioutil.ReadFile(f)
		assert.NoError(err)
		tpl := template.Must(template.New(f).Parse(string(s)))
		err = test.applier.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}
