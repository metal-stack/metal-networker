package netconf

import (
	"fmt"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"

	"github.com/metal-pod/v"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

// TplNftablesV4 defines the name of the template to render nftables configuration.
const (
	TplNftablesV4 = "nftrules.v4.tpl"
	TplNftablesV6 = "nftrules.v6.tpl"
)

type (
	// NftablesData represents the information required to render nftables configuration.
	NftablesData struct {
		Comment string
		SNAT    []SNAT
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
func NewNftablesConfigApplier(kb KnowledgeBase, validator network.Validator) network.Applier {
	data := NftablesData{
		Comment: fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n"+
			"# Do not edit.", kb.Machineuuid, v.V.String()),
		SNAT: getSNAT(kb),
	}
	return network.NewNetworkApplier(data, validator, nil)
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
