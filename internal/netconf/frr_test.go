package netconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileFrrConf(t *testing.T) {
	tests := []FileRenderInfo{
		{
			input:            "testdata/firewall.yaml",
			expectedOutput:   "testdata/frr.conf.firewall",
			configuratorType: Firewall,
			tpl:              TplFirewallFRR,
			newApplierFunc:   NewFrrConfigApplier,
		},
		{
			input:            "testdata/machine.yaml",
			expectedOutput:   "testdata/frr.conf.machine",
			configuratorType: Machine,
			tpl:              TplMachineFRR,
			newApplierFunc:   NewFrrConfigApplier,
		},
	}
	renderFilesAndVerifyExpectations(t, tests)
}

func TestFRRValidator_Validate(t *testing.T) {
	assert := assert.New(t)

	validator := FRRValidator{}
	actual := validator.Validate()
	assert.NotNil(actual)
	assert.NotNil(actual.Error())
}
