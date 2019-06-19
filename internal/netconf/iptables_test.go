package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestCompileRules(t *testing.T) {
	assert := assert.New(t)
	expected, err := ioutil.ReadFile("testdata/rules.v4")
	assert.NoError(err)

	kb := NewKnowledgeBase("testdata/install.yaml")
	assert.NoError(err)

	a := NewIptablesConfig(kb, "")
	b := bytes.Buffer{}

	f := TplIptables
	s, err := ioutil.ReadFile(f)
	assert.NoError(err)
	tpl := template.Must(template.New(f).Parse(string(s)))
	err = a.Applier.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
