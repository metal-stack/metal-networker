package netconf

const (
	// AllZerosCIDR represents a CIDR notation that matches all addresses in the IPv4 address space.
	AllZerosCIDR = "0.0.0.0/0"
	// Permit defines an access policy that allows access.
	Permit AccessPolicy = iota
	// Deny defines an access policy that forbids access.
	Deny
)

type (
	// AccessPolicy is a type that represents a policy to manage access roles.
	AccessPolicy int

	// Identity represents an object's identity.
	Identity struct {
		Comment string
		ID      int
	}

	// Loopback represents a loopback interface (lo).
	Loopback struct {
		Comment string
		IPs     []string
	}

	// VRF represents data required to render VRF information into frr.conf.
	VRF struct {
		Identity
		VNI            int
		ImportVRFNames []string
		IPPrefixLists  []IPPrefixList
		RouteMaps      []RouteMap
	}

	// RouteMap represents a route-map to permit or deny routes.
	RouteMap struct {
		Name    string
		Entries []string
		Policy  string
		Order   int
	}

	// IPPrefixList represents 'ip prefix-list' filtering mechanism to be used in combination with route-maps.
	IPPrefixList struct {
		Name string
		Spec string
	}

	// SVI represents a switched virtual interface.
	SVI struct {
		VlanID    int
		Comment   string
		Addresses []string
	}

	// VXLAN represents a VXLAN interface.
	VXLAN struct {
		Identity
		TunnelIP string
	}

	// EVPNIface represents the information required to render EVPN interfaces configuration.
	EVPNIface struct {
		VRF   VRF
		SVI   SVI
		VXLAN VXLAN
	}

	// Bridge represents a network bridge.
	Bridge struct {
		Ports string
		Vids  string
	}
)

func (p AccessPolicy) String() string {
	switch p {
	case Permit:
		return "permit"
	case Deny:
		return "deny"
	}

	return "undefined"
}
