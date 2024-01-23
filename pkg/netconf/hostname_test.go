package netconf

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNameHostname(t *testing.T) {
	expected, err := os.ReadFile("testdata/hostname")
	require.NoError(t, err)

	log := slog.Default()
	kb, err := New(log, "testdata/firewall.yaml")
	require.NoError(t, err)

	a := newHostnameApplier(*kb, "")
	b := bytes.Buffer{}

	tpl := MustParseTpl(tplHostname)
	err = a.Render(&b, *tpl)
	require.NoError(t, err)
	assert.Equal(t, string(expected), b.String())
}
