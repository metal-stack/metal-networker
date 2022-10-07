package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNewHostsApplier(t *testing.T) {
	assert := assert.New(t)

	expected, err := os.ReadFile("testdata/hosts")
	assert.NoError(err)

	log := zaptest.NewLogger(t).Sugar()
	kb, err := New(log, "testdata/firewall.yaml")
	assert.NoError(err)
	a := newHostsApplier(*kb, "")
	b := bytes.Buffer{}

	tpl := mustParseTpl(tplHosts)
	err = a.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
