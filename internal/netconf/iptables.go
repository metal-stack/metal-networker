package netconf

import (
	"fmt"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"

	"github.com/metal-pod/v"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

// TplIptables defines the name of the template to render iptables configuration.
const TplIptables = "rules.v4.tpl"

type (
	// IptablesData represents the information required to render iptables configuration.
	IptablesData struct {
		Comment string
		SNAT    []SNAT
	}

	// SNAT holds the information required to configure Source NAT.
	SNAT struct {
		Comment      string
		OutInterface string
		SourceSpecs  []string
	}

	// IptablesValidator can validate configuration for network interfaces.
	IptablesValidator struct {
		path string
	}
)

// NewIptablesConfigApplier constructs a new instance of this type.
func NewIptablesConfigApplier(kb KnowledgeBase, tmpFile string) network.Applier {
	data := IptablesData{
		Comment: fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n"+
			"# Do not edit.", kb.Machineuuid, v.V.String()),
		SNAT: getSNAT(kb),
	}
	validator := IptablesValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
}

func getSNAT(kb KnowledgeBase) []SNAT {
	var result []SNAT
	primary := kb.getPrimaryNetwork()

	networks := kb.GetNetworks(Primary, External)
	for _, n := range networks {
		if !n.Nat {
			continue
		}

		var sources []string
		sources = append(sources, primary.Prefixes...)
		cmt := fmt.Sprintf("snat (networkid: %s)", n.Networkid)
		svi := fmt.Sprintf("vlan%d", n.Vrf)

		s := SNAT{
			Comment:      cmt,
			OutInterface: svi,
			SourceSpecs:  sources,
		}
		result = append(result, s)
	}
	return result
}

// Validate validates network interfaces configuration.
func (v IptablesValidator) Validate() error {
	log.Infof("running 'ifup --syntax-check --all --interfaces %s to validate changes.'", v.path)
	return exec.NewVerboseCmd("iptables-restore", "--test", "--verbose", v.path).Run()
}
