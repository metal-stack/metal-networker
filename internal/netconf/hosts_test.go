package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestNewHostsApplier(t *testing.T) {
	assert := assert.New(t)

	expected, err := ioutil.ReadFile("testdata/hosts")
	assert.NoError(err)

	kb := NewKnowledgeBase("testdata/firewall.yaml")
	assert.NoError(err)
	a := NewHostsApplier(kb, "")
	b := bytes.Buffer{}

	s, err := ioutil.ReadFile(TplHosts)
	assert.NoError(err)
	tpl := template.Must(template.New(TplHosts).Parse(string(s)))
	err = a.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
