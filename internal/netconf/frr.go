package netconf

import (
	"fmt"

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
	d.ASN = kb.mustGetUnderlay().Asn
	d.Comment = versionHeader(kb.Machineuuid)
	d.FRRVersion = FRRVersion
	d.Hostname = kb.Hostname
	d.RouterID = kb.mustGetUnderlay().Ips[0]
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
	for _, n := range kb.Networks {
		// VRF BGP Instances are configured for tenant network (primary) and all external networks
		// (non underlay) to enable traffic from tenant network into external networks and vice versa.
		if n.Underlay {
			continue
		}
		vrf := VRF{}
		vrf.ID = n.Vrf
		vrf.VNI = n.Vrf
		if n.Primary {
			// The primary vrf contains a static route leak into VRF's of external networks.
			// In addition to this the primary vrf announces a default route to ask clients to route all traffic destined
			// to external networks to here.
			vrf.RouteImports = getRouteImportsPrimary(kb) // import destination prefixes
		} else {
			vrf.RouteImports = getRouteImportsNonPrimary(kb) // import destination prefixes
		}
		result = append(result, vrf)
	}
	return result
}

func getRouteImportsPrimary(kb KnowledgeBase) []RouteImport {
	result := []RouteImport{}
	for _, n := range kb.Networks {
		// The primary and underlay networks are not targets to route external traffic to.
		if n.Primary || n.Underlay {
			continue
		}
		if len(n.Destinationprefixes) == 0 {
			continue
		}
		ri := RouteImport{SourceVRF: fmt.Sprintf("vrf%d", n.Vrf), AllowedImportPrefixes: n.Destinationprefixes}
		result = append(result, ri)
	}
	return result
}

func getRouteImportsNonPrimary(kb KnowledgeBase) []RouteImport {
	result := []RouteImport{}
	primary := kb.mustGetPrimary()
	allowed := []string{}
	allowed = append(allowed, primary.Prefixes...)
	for _, n := range kb.Networks {
		// The primary and underlay networks are not targets to route external traffic to.
		if n.Primary || n.Underlay {
			continue
		}
		allowed = append(allowed, n.Prefixes...)
	}
	if len(allowed) == 0 {
		return result
	}

	allowedWith32 := []string{}
	for _, a := range allowed {
		allowedWith32 = append(allowedWith32, a+" le 32")
	}
	ri := RouteImport{SourceVRF: fmt.Sprintf("vrf%d", primary.Vrf), AllowedImportPrefixes: allowedWith32}
	result = append(result, ri)
	return result
}
