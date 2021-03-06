package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-networker/pkg/exec"
	"inet.af/netaddr"

	"github.com/metal-stack/v"

	"github.com/metal-stack/metal-networker/pkg/net"

	mn "github.com/metal-stack/metal-lib/pkg/net"
)

const (
	// TplNftables defines the name of the template to render nftables configuration.
	TplNftables = "nftrules.tpl"
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
		SourceSpecs  []SourceSpec
	}

	SourceSpec struct {
		AddressFamily string
		Source        string
	}
	// NftablesValidator can validate configuration for nftables rules.
	NftablesValidator struct {
		path string
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

func isDMZNetwork(n models.V1MachineNetwork) bool {
	return *n.Networktype == mn.PrivateSecondaryShared && containsDefaultRoute(n.Destinationprefixes)
}

func getSNAT(kb KnowledgeBase) []SNAT {
	var result []SNAT

	private := kb.getPrivatePrimaryNetwork()
	networks := kb.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared, mn.PrivateSecondaryShared, mn.External)

	privatePfx := private.Prefixes
	for _, n := range kb.Networks {
		if isDMZNetwork(n) {
			privatePfx = append(privatePfx, n.Prefixes...)
		}
	}

	for _, n := range networks {
		if n.Nat != nil && !*n.Nat {
			continue
		}

		var sources []SourceSpec
		cmt := fmt.Sprintf("snat (networkid: %s)", *n.Networkid)
		svi := fmt.Sprintf("vlan%d", *n.Vrf)

		for _, p := range privatePfx {
			ipprefix, err := netaddr.ParseIPPrefix(p)
			if err != nil {
				continue
			}
			af := "ip"
			if ipprefix.IP.Is6() {
				af = "ip6"
			}
			sspec := SourceSpec{
				Source:        p,
				AddressFamily: af,
			}
			sources = append(sources, sspec)
		}
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
func (v NftablesValidator) Validate() error {
	log.Infof("running 'nft --check --file %s' to validate changes.", v.path)
	return exec.NewVerboseCmd("nft", "--check", "--file", v.path).Run()
}
