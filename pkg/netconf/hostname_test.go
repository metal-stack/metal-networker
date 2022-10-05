package netconf

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNameHostname(t *testing.T) {
	assert := assert.New(t)
	expected, err := os.ReadFile("testdata/hostname")
	assert.NoError(err)

	log := zaptest.NewLogger(t).Sugar()
	kb, err := New(log, "testdata/firewall.yaml")
	assert.NoError(err)

	a := newHostnameApplier(*kb, "")
	b := bytes.Buffer{}

	tpl := mustParseTpl(tplHostname)
	err = a.Render(&b, *tpl)
	assert.NoError(err)
	assert.Equal(string(expected), b.String())
}
