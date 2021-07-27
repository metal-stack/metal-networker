package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHostsApplier(t *testing.T) {
	assert := assert.New(t)

	expected, err := os.ReadFile("testdata/hosts")
	assert.NoError(err)

	kb := NewKnowledgeBase("testdata/firewall.yaml")
	assert.NoError(err)
	a := NewHostsApplier(kb, "")
	b := bytes.Buffer{}

	tpl := mustParseTpl(TplHosts)
	err = a.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
