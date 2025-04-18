package netconf

import (
	"fmt"
	"log/slog"
	"net/netip"

	"github.com/Masterminds/semver/v3"
	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"github.com/metal-stack/metal-networker/pkg/exec"
	"github.com/metal-stack/metal-networker/pkg/net"
)

const (
	// FRRVersion holds a string that is used in the frr.conf to define the FRR version.
	FRRVersion = "8.5"
	// TplFirewallFRR defines the name of the template to render FRR configuration to a 'firewall'.
	TplFirewallFRR = "frr.firewall.tpl"
	// TplMachineFRR defines the name of the template to render FRR configuration to a 'machine'.
	TplMachineFRR = "frr.machine.tpl"
	// IPPrefixListSeqSeed specifies the initial value for prefix lists sequence number.
	IPPrefixListSeqSeed = 100
	// IPPrefixListNoExportSuffix defines the suffix to use for private IP ranges that must not be exported.
	IPPrefixListNoExportSuffix = "-no-export"
	// RouteMapOrderSeed defines the initial value for route-map order.
	RouteMapOrderSeed = 10
	// AddressFamilyIPv4 is the name for this address family for the routing daemon.
	AddressFamilyIPv4 = "ip"
	// AddressFamilyIPv6 is the name for this address family for the routing daemon.
	AddressFamilyIPv6 = "ipv6"
)

type (
	// CommonFRRData contains attributes that are common to FRR configuration of all kind of bare metal servers.
	CommonFRRData struct {
		ASN        int64
		Comment    string
		FRRVersion string
		Hostname   string
		RouterID   string
	}

	// MachineFRRData contains attributes required to render frr.conf of bare metal servers that function as 'machine'.
	MachineFRRData struct {
		CommonFRRData
	}

	// FirewallFRRData contains attributes required to render frr.conf of bare metal servers that function as 'firewall'.
	FirewallFRRData struct {
		CommonFRRData
		VRFs []VRF
	}

	// frrValidator validates the frr.conf to apply.
	frrValidator struct {
		path string
		log  *slog.Logger
	}

	// AddressFamily is the address family for the routing daemon.
	AddressFamily string
)

// NewFrrConfigApplier constructs a new Applier of the given type of Bare Metal.
func NewFrrConfigApplier(kind BareMetalType, c config, tmpFile string, frrVersion *semver.Version) net.Applier {
	var data any

	switch kind {
	case Firewall:
		net := c.getUnderlayNetwork()
		data = FirewallFRRData{
			CommonFRRData: CommonFRRData{
				FRRVersion: FRRVersion,
				Hostname:   c.Hostname,
				Comment:    versionHeader(c.MachineUUID),
				ASN:        *net.Asn,
				RouterID:   routerID(net),
			},
			VRFs: assembleVRFs(c, frrVersion),
		}
	case Machine:
		net := c.getPrivatePrimaryNetwork()
		data = MachineFRRData{
			CommonFRRData: CommonFRRData{
				FRRVersion: FRRVersion,
				Hostname:   c.Hostname,
				Comment:    versionHeader(c.MachineUUID),
				ASN:        *net.Asn,
				RouterID:   routerID(net),
			},
		}
	default:
		c.log.Error("unknown kind of bare metal", "kind", kind)
		panic(fmt.Errorf("unknown kind %v", kind))
	}

	validator := frrValidator{
		path: tmpFile,
		log:  c.log,
	}

	return net.NewNetworkApplier(data, validator, net.NewDBusReloader("frr.service"))
}

// routerID will calculate the bgp router-id which must only be specified in the ipv6 range.
// returns 0.0.0.0 for erroneous ip addresses and 169.254.255.255 for ipv6
// TODO prepare machine allocations with ipv6 primary address and tests
func routerID(net *models.V1MachineNetwork) string {
	if len(net.Ips) < 1 {
		return "0.0.0.0"
	}
	ip, err := netip.ParseAddr(net.Ips[0])
	if err != nil {
		return "0.0.0.0"
	}
	if ip.Is4() {
		return ip.String()
	}
	return "169.254.255.255"
}

// Validate can be used to run validation on FRR configuration using vtysh.
func (v frrValidator) Validate() error {
	vtysh := fmt.Sprintf("vtysh --dryrun --inputfile %s", v.path)
	v.log.Info("validate changes", "command", vtysh)

	return exec.NewVerboseCmd("bash", "-c", vtysh, v.path).Run()
}

func assembleVRFs(kb config, frrVersion *semver.Version) []VRF {
	var (
		result []VRF
		frr    *FRR
	)
	if frrVersion != nil {
		frr = &FRR{
			Major: frrVersion.Major(),
			Minor: frrVersion.Minor(),
		}
	}

	networks := kb.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared, mn.PrivateSecondaryShared, mn.External)
	for _, network := range networks {
		if network.Networktype == nil {
			continue
		}

		i := importRulesForNetwork(kb, network)
		vrf := VRF{
			Identity: Identity{
				ID: int(*network.Vrf),
			},
			VNI:            int(*network.Vrf),
			ImportVRFNames: i.ImportVRFs,
			IPPrefixLists:  i.prefixLists(),
			RouteMaps:      i.routeMaps(),
			FRRVersion:     frr,
		}
		result = append(result, vrf)
	}

	return result
}
