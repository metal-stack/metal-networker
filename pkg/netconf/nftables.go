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
		Input         Input
	}

	Input struct {
		InInterfaces []string
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
		Input:         getInput(c),
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

func getInput(c config) Input {
	input := Input{}
	networks := c.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared, mn.PrivateSecondaryShared)
	for _, n := range networks {
		input.InInterfaces = append(input.InInterfaces, fmt.Sprintf("vrf%d", *n.Vrf))
	}
	return input
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
			af, err := getAddressFamily(p)
			if err != nil {
				continue
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
		DAddr:        "@proxy_dns_servers",
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
		egressRules           = []string{"# egress rules specified during firewall creation"}
		ingressRules          = []string{"# ingress rules specified during firewall creation"}
		inputInterfaces       = getInput(c)
		quotedInputInterfaces []string
	)
	for _, i := range inputInterfaces.InInterfaces {
		quotedInputInterfaces = append(quotedInputInterfaces, "\""+i+"\"")
	}

	for _, r := range c.FirewallRules.Egress {
		ports := make([]string, len(r.Ports))
		for i, v := range r.Ports {
			ports[i] = strconv.Itoa(int(v))
		}
		for _, daddr := range r.To {
			af, err := getAddressFamily(daddr)
			if err != nil {
				continue
			}
			egressRules = append(egressRules,
				fmt.Sprintf("iifname { %s } %s daddr %s %s dport { %s } counter accept comment %q", strings.Join(quotedInputInterfaces, ","), af, daddr, strings.ToLower(r.Protocol), strings.Join(ports, ","), r.Comment))
		}
	}

	privatePrimaryNetwork := c.getPrivatePrimaryNetwork()
	outputInterfacenames := ""
	if privatePrimaryNetwork != nil && privatePrimaryNetwork.Vrf != nil {
		outputInterfacenames = fmt.Sprintf("oifname { \"vrf%d\", \"vni%d\", \"vlan%d\" }", *privatePrimaryNetwork.Vrf, *privatePrimaryNetwork.Vrf, *privatePrimaryNetwork.Vrf)
	}

	for _, r := range c.FirewallRules.Ingress {
		ports := make([]string, len(r.Ports))
		for i, v := range r.Ports {
			ports[i] = strconv.Itoa(int(v))
		}
		destinationSpec := ""
		if len(r.To) > 0 {
			destinationSpec = fmt.Sprintf("ip daddr { %s }", strings.Join(r.To, ", "))
		} else if outputInterfacenames != "" {
			destinationSpec = outputInterfacenames
		} else {
			c.log.Warn("no to address specified but not private primary network present, skipping this rule", "rule", r)
			continue
		}

		for _, saddr := range r.From {
			af, err := getAddressFamily(saddr)
			if err != nil {
				continue
			}
			ingressRules = append(ingressRules, fmt.Sprintf("%s %s saddr %s %s dport { %s } counter accept comment %q", destinationSpec, af, saddr, strings.ToLower(r.Protocol), strings.Join(ports, ","), r.Comment))
		}
	}
	return FirewallRules{
		Egress:  egressRules,
		Ingress: ingressRules,
	}
}

func getAddressFamily(p string) (string, error) {
	prefix, err := netip.ParsePrefix(p)
	if err != nil {
		return "", err
	}
	family := "ip"
	if prefix.Addr().Is6() {
		family = "ip6"
	}
	return family, nil
}

// Validate validates network interfaces configuration.
func (v NftablesValidator) Validate() error {
	v.log.Info("running 'nft --check --file' to validate changes.", "file", v.path)
	return exec.NewVerboseCmd("nft", "--check", "--file", v.path).Run()
}
