package netconf

import (
	"fmt"
	"net"
	"strings"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

const (
	// FRRVersion holds a string that is used in the frr.conf to define the FRR version.
	FRRVersion = "7.0"
	// TplFirewallFRR defines the name of the template to render FRR configuration to a 'firewall'.
	TplFirewallFRR = "frr.firewall.tpl"
	// TplMachineFRR defines the name of the template to render FRR configuration to a 'machine'.
	TplMachineFRR = "frr.machine.tpl"
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
		LocalBGPIP string
	}

	// FirewallFRRData contains attributes required to render frr.conf of bare metal servers that function as 'firewall'.
	FirewallFRRData struct {
		CommonFRRData
		VRFs []VRF
	}

	// VRF represents data required to render VRF information into frr.conf.
	VRF struct {
		ID           int
		VNI          int
		RouteImports []RouteImport
	}

	// RouteImport represents data to apply for dynamic route leak configuration.
	RouteImport struct {
		SourceVRF             string
		AllowedImportPrefixes []string
	}

	// FRRValidator validates the frr.conf to apply.
	FRRValidator struct {
		path string
	}
)

// NewFrrConfigApplier constructs a new Applier of the given type of Bare Metal.
func NewFrrConfigApplier(kind BareMetalType, kb KnowledgeBase, tmpFile string) network.Applier {
	var data interface{}

	switch kind {
	case Firewall:
		net := kb.getUnderlayNetwork()
		common := newCommonFRRData(net, kb)
		data = FirewallFRRData{CommonFRRData: common, VRFs: getVRFs(kb)}
	case Machine:
		net := kb.getPrimaryNetwork()
		common := newCommonFRRData(net, kb)
		localBGPIP := getLocalBGPIP(kb)
		data = MachineFRRData{common, localBGPIP}
	default:
		log.Fatalf("unknown kind of bare metal: %v", kind)
	}

	validator := FRRValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
}

func getLocalBGPIP(kb KnowledgeBase) string {
	primaryIPs := kb.getPrimaryNetwork().Ips
	ip := net.ParseIP(primaryIPs[0])
	ip[len(ip)-1] = 0
	return ip.String()
}

func newCommonFRRData(net Network, kb KnowledgeBase) CommonFRRData {
	return CommonFRRData{FRRVersion: FRRVersion, Hostname: kb.Hostname, Comment: versionHeader(kb.Machineuuid),
		ASN: net.Asn, RouterID: net.Ips[0]}
}

// Validate can be used to run validation on FRR configuration using vtysh.
func (v FRRValidator) Validate() error {
	vtysh := fmt.Sprintf("vtysh --dryrun --inputfile %s", v.path)
	log.Infof("running '%s' to validate changes.'", vtysh)
	return exec.NewVerboseCmd("bash", "-c", vtysh, v.path).Run()
}

func getVRFs(kb KnowledgeBase) []VRF {
	var result []VRF
	networks := kb.GetNetworks(Primary, External)
	for _, n := range networks {
		vrf := VRF{ID: n.Vrf, VNI: n.Vrf}
		// Between VRFs we use dynamic route leak to import the desired prefixes
		if n.Primary {
			// Import routes to reach out from primary network into external networks.
			vrf.RouteImports = getImportsFromExternalNetworks(kb)
		} else {
			// Import routes to reach out from an external network into primary and other external networks.
			vrf.RouteImports = getImportsFromPrimaryNetwork(kb, n)
		}
		result = append(result, vrf)
	}
	return result
}

func getImportsFromExternalNetworks(kb KnowledgeBase) []RouteImport {
	var result []RouteImport
	networks := kb.GetNetworks(External)
	for _, n := range networks {
		isEmptyDestination := len(n.Destinationprefixes) == 0
		if isEmptyDestination {
			continue
		}
		var allowed []string
		for _, dp := range n.Destinationprefixes {
			if strings.HasSuffix(dp, "/0") {
				allowed = append(allowed, dp)
			} else {
				// Prefix list will be applied if the prefix length is less than or equal to the le prefix length.
				allowed = append(allowed, dp+" le 32")
			}
		}
		ri := RouteImport{SourceVRF: fmt.Sprintf("vrf%d", n.Vrf), AllowedImportPrefixes: allowed}
		result = append(result, ri)
	}
	return result
}

func getImportsFromPrimaryNetwork(kb KnowledgeBase, n Network) []RouteImport {
	var result []RouteImport
	var prefixes []string
	primary := kb.getPrimaryNetwork()
	// considers the case: machine's associated IP within the primary network
	prefixes = append(prefixes, primary.Prefixes...)
	// considers the case: machine's associated IP within the given network
	prefixes = append(prefixes, n.Prefixes...)
	if len(prefixes) == 0 {
		return result
	}

	var allowed []string
	for _, p := range prefixes {
		allowed = append(allowed, p+" le 32")
	}
	ri := RouteImport{SourceVRF: fmt.Sprintf("vrf%d", primary.Vrf), AllowedImportPrefixes: allowed}
	result = append(result, ri)
	return result
}
