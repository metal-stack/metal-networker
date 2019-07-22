package netconf

import (
	"testing"
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
