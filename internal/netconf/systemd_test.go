package netconf

import (
	"bytes"
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/metal-stack/metal-networker/pkg/net"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSystemdLinkConfig(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		expectedOutput   string
		configuratorType BareMetalType
		tpl              string
		nicIndex         int
		machineUUID      string
	}{
		{expectedOutput: "testdata/lan0.machine.link",
			configuratorType: Machine,
			tpl:              TplSystemdLink,
			machineUUID:      "e0ab02d2-27cd-5a5e-8efc-080ba80cf258",
			nicIndex:         0},
		{expectedOutput: "testdata/lan1.machine.link",
			configuratorType: Machine,
			tpl:              TplSystemdLink,
			machineUUID:      "e0ab02d2-27cd-5a5e-8efc-080ba80cf258",
			nicIndex:         1},
		{expectedOutput: "testdata/lan0.firewall.link",
			configuratorType: Firewall,
			tpl:              TplSystemdLink,
			machineUUID:      "e0ab02d2-27cd-5a5e-8efc-080ba80cf258",
			nicIndex:         0},
		{expectedOutput: "testdata/lan1.firewall.link",
			configuratorType: Firewall,
			tpl:              TplSystemdLink,
			machineUUID:      "e0ab02d2-27cd-5a5e-8efc-080ba80cf258",
			nicIndex:         1},
	}

	for _, t := range tests {
		expected, err := ioutil.ReadFile(t.expectedOutput)
		assert.NoError(err)

		nic := NewKnowledgeBase("testdata/firewall.yaml", zap.NewNop().Sugar()).Nics[t.nicIndex]
		assert.NoError(err)
		a := NewSystemdLinkApplier(t.configuratorType, t.machineUUID, t.nicIndex, nic, "", zap.NewNop().Sugar())
		b := bytes.Buffer{}

		s, err := ioutil.ReadFile(t.tpl)
		assert.NoError(err)
		tpl := template.Must(template.New(t.tpl).Parse(string(s)))
		err = a.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}

func TestNewSystemdNetworkConfig(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		expectedOutput   string
		configuratorType BareMetalType
		tpl              string
		nicIndex         int
		machineUUID      string
		configFunc       func(machineUUID string, nicIndex int, tmpFile string, log *zap.SugaredLogger) net.Applier
	}{
		{expectedOutput: "testdata/lan0.network",
			configuratorType: Machine,
			tpl:              TplSystemdNetwork,
			nicIndex:         0,
			machineUUID:      "e0ab02d2-27cd-5a5e-8efc-080ba80cf258",
			configFunc:       NewSystemdNetworkApplier},
		{expectedOutput: "testdata/lan1.network",
			configuratorType: Machine,
			tpl:              TplSystemdNetwork,
			nicIndex:         1,
			machineUUID:      "e0ab02d2-27cd-5a5e-8efc-080ba80cf258",
			configFunc:       NewSystemdNetworkApplier},
	}

	for _, t := range tests {
		expected, err := ioutil.ReadFile(t.expectedOutput)
		assert.NoError(err)

		assert.NoError(err)
		a := t.configFunc(t.machineUUID, t.nicIndex, "", zap.NewNop().Sugar())
		b := bytes.Buffer{}

		s, err := ioutil.ReadFile(t.tpl)
		assert.NoError(err)
		tpl := template.Must(template.New(t.tpl).Parse(string(s)))
		err = a.Render(&b, *tpl)
		assert.NoError(err)
		assert.Equal(string(expected), b.String())
	}
}
