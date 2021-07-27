package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameHostname(t *testing.T) {
	assert := assert.New(t)
	expected, err := os.ReadFile("testdata/hostname")
	assert.NoError(err)

	kb := NewKnowledgeBase("testdata/firewall.yaml")
	assert.NoError(err)

	a := NewHostnameApplier(kb, "")
	b := bytes.Buffer{}

	tpl := mustParseTpl(TplHostname)
	err = a.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
