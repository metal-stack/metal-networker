package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"text/template"
)

func TestCompileInterfaces(t *testing.T) {
	assert := assert.New(t)
	expected, err := ioutil.ReadFile("test-data/interfaces")
	assert.NoError(err)

	kb, err := NewKnowledgeBase("test-data/install.yaml")
	assert.NoError(err)

	a := NewIfacesConfig(*kb, "")
	b := bytes.Buffer{}

	f := "test-data/interfaces.tpl"
	s, err := ioutil.ReadFile(f)
	tpl := template.Must(template.New(f).Parse(string(s)))
	err = a.Applier.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
