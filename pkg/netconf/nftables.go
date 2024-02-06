package netconf

import (
	"fmt"
	"log/slog"
	"net/netip"
	"strconv"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"

	"github.com/metal-stack/metal-networker/pkg/exec"
	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// TplNftables defines the name of the template to render nftables configuration.
	TplNftables     = "nftrules.tpl"
	dnsPort         = "domain"
	nftablesService = "nftables.service"
	systemctlBin    = "/bin/systemctl"

	// Set up additional conntrack zone for DNS traffic.
	// There was a problem that duplicate packets were registered by conntrack
	// when packet was leaking from private VRF to the internet VRF.
	// Isolating traffic to special zone solves the problem.
	// Zone number(3) was obtained by experiments.
	dnsProxyZone = "3"
)

type (
	// NftablesData represents the information required to render nftables configuration.
	NftablesData struct {
		Comment       string
		SNAT          []SNAT
		DNSProxyDNAT  DNAT
		VPN           bool
		ForwardPolicy string
		FirewallRules FirewallRules
	}

	FirewallRules struct {
		Egress  []string
		Ingress []string
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
		DAddr        string
		Port         string
		Zone         string
		DestSpec     AddrSpec
	}

	AddrSpec struct {
		AddressFamily string
		Address       string
	}

	// NftablesValidator can validate configuration for nftables rules.
	NftablesValidator struct {
		path string
		log  *slog.Logger
	}

	NftablesReloader struct{}
)

// newNftablesConfigApplier constructs a new instance of this type.
func newNftablesConfigApplier(c config, validator net.Validator, enableDNSProxy bool, forwardPolicy ForwardPolicy) net.Applier {
	data := NftablesData{
		Comment:       versionHeader(c.MachineUUID),
		SNAT:          getSNAT(c, enableDNSProxy),
		ForwardPolicy: string(forwardPolicy),
		FirewallRules: getFirewallRules(c),
	}

	if enableDNSProxy {
		data.DNSProxyDNAT = getDNSProxyDNAT(c, dnsPort, dnsProxyZone)
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

func getDNSProxyDNAT(c config, port, zone string) DNAT {
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
		DAddr:        "@public_dns_servers",
		Port:         port,
		Zone:         zone,
		DestSpec: AddrSpec{
			AddressFamily: af,
			Address:       n.Ips[0],
		},
	}
}

func getFirewallRules(c config) FirewallRules {
	if c.FirewallRules == nil {
		return FirewallRules{}
	}
	var (
		egressRules  = []string{"# egress rules specified during firewall creation"}
		ingressRules = []string{"# ingress rules specified during firewall creation"}
	)
	for _, r := range c.FirewallRules.Egress {
		ports := make([]string, len(r.Ports))
		for i, v := range r.Ports {
			ports[i] = strconv.Itoa(int(v))
		}
		for _, daddr := range r.ToCidrs {
			prefix, err := netip.ParsePrefix(daddr)
			if err != nil {
				continue
			}
			family := "ip"
			if prefix.Addr().Is6() {
				family = "ip6"
			}
			egressRules = append(egressRules,
				fmt.Sprintf("ip saddr { 10.0.0.0/8 } %s daddr %s %s dport { %s } accept comment %q", family, daddr, r.Protocol, strings.Join(ports, ","), r.Comment))
		}
	}
	for _, r := range c.FirewallRules.Ingress {
		ports := make([]string, len(r.Ports))
		for i, v := range r.Ports {
			ports[i] = strconv.Itoa(int(v))
		}
		for _, saddr := range r.FromCidrs {
			prefix, err := netip.ParsePrefix(saddr)
			if err != nil {
				continue
			}
			family := "ip"
			if prefix.Addr().Is6() {
				family = "ip6"
			}
			ingressRules = append(ingressRules, fmt.Sprintf("ip daddr { 10.0.0.0/8 } %s saddr %s %s dport { %s } accept comment %q", family, saddr, r.Protocol, strings.Join(ports, ","), r.Comment))
		}
	}
	return FirewallRules{
		Egress:  egressRules,
		Ingress: ingressRules,
	}
}

// Validate validates network interfaces configuration.
func (v NftablesValidator) Validate() error {
	v.log.Info("running 'nft --check --file' to validate changes.", "file", v.path)
	return exec.NewVerboseCmd("nft", "--check", "--file", v.path).Run()
}
