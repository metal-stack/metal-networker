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
	// TplFRR defines the name of the template to render FRR configuration.
	TplFRR = "frr.tpl"
)

// FRRData represents the information required to render frr.conf.
type FRRData struct {
	FRRVersion string
	ASN        int64
	Comment    string
	Hostname   string
	RouterID   string
	VRFs       []VRF
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

// FRRConfig represents a thing to apply changes to frr.conf.
type FRRConfig struct {
	Applier network.Applier
}

// NewFRRConfig constructs a new instance of this type.
func NewFRRConfig(kb KnowledgeBase, tmpFile string) FRRConfig {
	d := FRRData{}
	d.ASN = kb.getUnderlayNetwork().Asn
	d.Comment = versionHeader(kb.Machineuuid)
	d.FRRVersion = FRRVersion
	d.Hostname = kb.Hostname
	d.RouterID = kb.getUnderlayNetwork().Ips[0]
	d.VRFs = getVRFs(kb)

	v := FRRValidator{tmpFile}
	a := network.NewNetworkApplier(d, v, nil)

	return FRRConfig{a}
}

// FRRValidator validates the frr.conf to apply.
type FRRValidator struct {
	path string
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
