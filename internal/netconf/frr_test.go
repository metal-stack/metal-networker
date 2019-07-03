package netconf

import (
	"testing"
)

func TestCompileFrrConf(t *testing.T) {
	tests := []FileRenderInfo{
		{expectedOutput: "testdata/frr.conf.firewall", configuratorType: Firewall, tpl: TplFirewallFRR,
			newApplierFunc: NewFrrConfigApplier},
		{expectedOutput: "testdata/frr.conf.machine", configuratorType: Machine, tpl: TplMachineFRR,
			newApplierFunc: NewFrrConfigApplier},
	}
	renderFilesAndVerifyExpectations(t, tests)
}
