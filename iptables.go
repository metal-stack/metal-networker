package main

import (
	"fmt"

	"github.com/metal-pod/v"

	"go.uber.org/zap"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

// IptablesConfig represents a thing to apply changes to iptables configuration.
type IptablesConfig struct {
	Applier network.Applier
	Log     zap.Logger
}

// IptablesData represents the information required to render iptables configuration.
type IptablesData struct {
	Comment string
	SNAT    []SNAT
}

// SNAT holds the information required to configure Source NAT.
type SNAT struct {
	Comment      string
	OutInterface string
	SourceSpecs  []string
}

// NewIptablesConfig constructs a new instance of this type.
func NewIptablesConfig(kb KnowledgeBase, tmpFile string) IptablesConfig {
	d := IptablesData{}
	d.Comment = fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n# Do not edit.",
		kb.Machineuuid, v.V.String())
	d.SNAT = getSNAT(kb)

	v := IptablesValidator{tmpFile}
	a := network.NewNetworkApplier(d, v, nil)
	return IptablesConfig{Applier: a}
}

func getSNAT(kb KnowledgeBase) []SNAT {
	var result []SNAT
	primary := kb.mustGetPrimary()

	for _, n := range kb.Networks {
		if n.Underlay || !n.Nat {
			continue
		}

		var sources []string
		sources = append(sources, primary.Prefixes...)
		cmt := fmt.Sprintf("snat (networkid: %s)", n.Networkid)
		svi := fmt.Sprintf("vlan%d", n.Vrf)

		s := SNAT{}
		s.Comment = cmt
		s.OutInterface = svi
		s.SourceSpecs = sources
		result = append(result, s)
	}
	return result
}

// IptablesValidator can validate configuration for network interfaces.
type IptablesValidator struct {
	path string
}

// Validate validates network interfaces configuration.
func (v IptablesValidator) Validate() error {
	log.Infof("running 'ifup --syntax-check --all --interfaces %s to validate changes.'", v.path)
	return NewVerboseCmd("iptables-restore", "--test", "--verbose", v.path).Run()
}
