package netconf

import (
	"fmt"
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

// CommonFRRData contains attributes that are common to FRR configuration of all kind of bare metal servers.
type CommonFRRData struct {
	ASN        int64
	Comment    string
	FRRVersion string
	Hostname   string
	RouterID   string
}

// MachineFRRData contains attributes required to render frr.conf of bare metal servers that function as 'machine'.
type MachineFRRData struct {
	CommonFRRData
}

// FirewallFRRData contains attributes required to render frr.conf of bare metal servers that function as 'firewall'.
type FirewallFRRData struct {
	CommonFRRData
	VRFs []VRF
}

// VRF represents data required to render VRF information into frr.conf.
type VRF struct {
	ID           int
	VNI          int
	RouteImports []RouteImport
}

// RouteImport represents data to apply for dynamic route leak configuration.
type RouteImport struct {
	SourceVRF             string
	AllowedImportPrefixes []string
}

// FRRValidator validates the frr.conf to apply.
type FRRValidator struct {
	path string
}

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
		data = MachineFRRData{common}
	default:
		log.Fatalf("unknown kind of bare metal: %v", kind)
	}

	validator := FRRValidator{tmpFile}
	return network.NewNetworkApplier(data, validator, nil)
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
	primary := kb.getPrimaryNetwork()
	networks := kb.GetNetworks(Primary, External)
	for _, n := range networks {
		vrf := VRF{ID: n.Vrf, VNI: n.Vrf}
		// Between VRFs we use dynamic route leak to import the desired prefixes
		if n.Primary {
			// Import routes to reach out from primary network into external networks.
			vrf.RouteImports = getRouteImportsIntoExternalNetworks(kb)
		} else {
			// Import routes to reach out from an external network into primary and other external networks.
			vrf.RouteImports = getRouteImportsInto(primary, n)
		}
		result = append(result, vrf)
	}
	return result
}

func getRouteImportsIntoExternalNetworks(kb KnowledgeBase) []RouteImport {
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

func getRouteImportsInto(primary, n Network) []RouteImport {
	var result []RouteImport
	var a []string
	a = append(a, primary.Prefixes...)
	a = append(a, n.Prefixes...)
	if len(a) == 0 {
		return result
	}

	var allowed []string
	for _, p := range a {
		allowed = append(allowed, p+" le 32")
	}
	ri := RouteImport{SourceVRF: fmt.Sprintf("vrf%d", primary.Vrf), AllowedImportPrefixes: allowed}
	result = append(result, ri)
	return result
}
