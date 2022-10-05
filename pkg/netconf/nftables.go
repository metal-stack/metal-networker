package netconf

import (
	"fmt"
	"net/netip"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"go.uber.org/zap"

	"github.com/metal-stack/metal-networker/pkg/exec"
	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// TplNftables defines the name of the template to render nftables configuration.
	TplNftables     = "nftrules.tpl"
	dnsPort         = "domain"
	nftablesService = "nftables.service"
	systemctlBin    = "/bin/systemctl"
)

type (
	// NftablesData represents the information required to render nftables configuration.
	NftablesData struct {
		Comment      string
		SNAT         []SNAT
		DNSProxyDNAT DNAT
		VPN          bool
	}

	// SNAT holds the information required to configure Source NAT.
	SNAT struct {
		Comment      string
		OutInterface string
		OutIntSpec   AddrSpec
		SourceSpecs  []AddrSpec
	}

	// DNAT holds the information required to configure DNAT.
	DNAT struct {
		Comment      string
		InInterfaces []string
		Port         string
		DestSpec     AddrSpec
	}

	AddrSpec struct {
		AddressFamily string
		Address       string
	}

	// NftablesValidator can validate configuration for nftables rules.
	NftablesValidator struct {
		path string
		log  *zap.SugaredLogger
	}

	NftablesReloader struct{}
)

// NewNftablesConfigApplier constructs a new instance of this type.
func NewNftablesConfigApplier(c config, validator net.Validator, enableDNSProxy bool) net.Applier {
	data := NftablesData{
		Comment: versionHeader(c.MachineUUID),
		SNAT:    getSNAT(c, enableDNSProxy),
	}

	if enableDNSProxy {
		data.DNSProxyDNAT = getDNSProxyDNAT(c, dnsPort)
	}

	if c.VPN != nil {
		data.VPN = true
	}

	return net.NewNetworkApplier(data, validator, &NftablesReloader{})
}

func (*NftablesReloader) Reload() error {
	return exec.NewVerboseCmd(systemctlBin, "reload", nftablesService).Run()
}

func isDMZNetwork(n *models.V1MachineNetwork) bool {
	return *n.Networktype == mn.PrivateSecondaryShared && containsDefaultRoute(n.Destinationprefixes)
}

func getSNAT(c config, enableDNSProxy bool) []SNAT {
	var result []SNAT

	private := c.getPrivatePrimaryNetwork()
	networks := c.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared, mn.PrivateSecondaryShared, mn.External)

	privatePfx := private.Prefixes
	for _, n := range c.Networks {
		if isDMZNetwork(n) {
			privatePfx = append(privatePfx, n.Prefixes...)
		}
	}

	var (
		defaultNetwork models.V1MachineNetwork
		defaultAF      string
	)
	defaultNetworkName, err := c.getDefaultRouteVRFName()
	if err == nil {
		defaultNetwork = *c.GetDefaultRouteNetwork()
		ip, _ := netip.ParseAddr(defaultNetwork.Ips[0])
		defaultAF = "ip"
		if ip.Is6() {
			defaultAF = "ip6"
		}
	}
	for _, n := range networks {
		if n.Nat != nil && !*n.Nat {
			continue
		}

		var sources []AddrSpec
		cmt := fmt.Sprintf("snat (networkid: %s)", *n.Networkid)
		svi := fmt.Sprintf("vlan%d", *n.Vrf)

		for _, p := range privatePfx {
			ipprefix, err := netip.ParsePrefix(p)
			if err != nil {
				continue
			}
			af := "ip"
			if ipprefix.Addr().Is6() {
				af = "ip6"
			}
			sspec := AddrSpec{
				Address:       p,
				AddressFamily: af,
			}
			sources = append(sources, sspec)
		}
		s := SNAT{
			Comment:      cmt,
			OutInterface: svi,
			SourceSpecs:  sources,
		}

		if enableDNSProxy && (vrfNameOf(n) == defaultNetworkName) {
			s.OutIntSpec = AddrSpec{
				AddressFamily: defaultAF,
				Address:       defaultNetwork.Ips[0],
			}
		}
		result = append(result, s)
	}

	return result
}

func getDNSProxyDNAT(c config, port string) DNAT {
	networks := c.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared, mn.PrivateSecondaryShared)
	svis := []string{}
	for _, n := range networks {
		svi := fmt.Sprintf("vlan%d", *n.Vrf)
		svis = append(svis, svi)
	}

	n := c.GetDefaultRouteNetwork()
	if n == nil {
		return DNAT{}
	}

	ip, _ := netip.ParseAddr(n.Ips[0])
	af := "ip"
	if ip.Is6() {
		af = "ip6"
	}
	return DNAT{
		Comment:      "dnat to dns proxy",
		InInterfaces: svis,
		Port:         port,
		DestSpec: AddrSpec{
			AddressFamily: af,
			Address:       n.Ips[0],
		},
	}
}

// Validate validates network interfaces configuration.
func (v NftablesValidator) Validate() error {
	v.log.Infof("running 'nft --check --file %s' to validate changes.", v.path)
	return exec.NewVerboseCmd("nft", "--check", "--file", v.path).Run()
}
