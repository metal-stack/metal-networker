package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestNameHostname(t *testing.T) {
	assert := assert.New(t)
	expected, err := ioutil.ReadFile("testdata/hostname")
	assert.NoError(err)

	kb := NewKnowledgeBase("testdata/install.yaml")
	assert.NoError(err)

	a := NewHostnameApplier(kb, "")
	b := bytes.Buffer{}

	f := TplHostname
	s, err := ioutil.ReadFile(f)
	assert.NoError(err)
	tpl := template.Must(template.New(f).Parse(string(s)))
	err = a.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
