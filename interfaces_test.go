package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"text/template"
)

const TemplateFile = "test-data/interfaces.tpl"

func TestCompile(t *testing.T) {
	assert := assert.New(t)
	expected, err := ioutil.ReadFile("test-data/interfaces")
	assert.NoError(err)

	kb, err := NewKnowledgeBase("test-data/install.yaml")
	assert.NoError(err)

	a := NewIfacesConfig(*kb, "")
	b := bytes.Buffer{}

	s, err := ioutil.ReadFile(TemplateFile)
	tpl := template.Must(template.New(TemplateFile).Parse(string(s)))
	err = a.Applier.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
