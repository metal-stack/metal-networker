package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/exec"

	"github.com/metal-stack/v"

	"github.com/metal-stack/metal-networker/pkg/net"
)

// TplNftablesV4 defines the name of the template to render nftables configuration.
const (
	TplNftablesV4 = "rules.v4.tpl"
	TplNftablesV6 = "rules.v6.tpl"
)

type (
	// NftablesData represents the information required to render nftables configuration.
	NftablesData struct {
		Comment string
		SNAT    []SNAT
	}

	// SNAT holds the information required to configure Source NAT.
	SNAT struct {
		Comment      string
		OutInterface string
		SourceSpecs  []string
	}

	// NftablesValidator can validate configuration for nftables rules.
	NftablesValidator struct {
		path string
	}

	// NftablesV4Validator can validate configuration for ipv4 nftables rules.
	NftablesV4Validator struct {
		NftablesValidator
	}

	// NftablesV6Validator can validate configuration for ipv6 nftables rules.
	NftablesV6Validator struct {
		NftablesValidator
	}
)

// NewNftablesConfigApplier constructs a new instance of this type.
func NewNftablesConfigApplier(kb KnowledgeBase, validator net.Validator) net.Applier {
	data := NftablesData{
		Comment: fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n"+
			"# Do not edit.", kb.Machineuuid, v.V.String()),
		SNAT: getSNAT(kb),
	}

	return net.NewNetworkApplier(data, validator, nil)
}

func getSNAT(kb KnowledgeBase) []SNAT {
	var result []SNAT

	private := kb.getPrivatePrimaryNetwork()
	networks := kb.GetNetworks(PrivatePrimaryUnshared, PrivatePrimaryShared, PrivateSecondaryShared, External)

	for _, n := range networks {
		if n.Nat != nil && !*n.Nat {
			continue
		}

		var sources []string
		sources = append(sources, private.Prefixes...)
		cmt := fmt.Sprintf("snat (networkid: %s)", *n.Networkid)
		svi := fmt.Sprintf("vlan%d", *n.Vrf)

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
func (v NftablesV4Validator) Validate() error {
	log.Infof("running 'nft --check --file %s' to validate changes.", v.path)
	return exec.NewVerboseCmd("nft", "--check", "--file", v.path).Run()
}

// Validate validates network interfaces configuration.
func (v NftablesV6Validator) Validate() error {
	log.Infof("running 'nft --check --file %s' to validate changes.", v.path)
	return exec.NewVerboseCmd("nft", "--check", "--file", v.path).Run()
}
