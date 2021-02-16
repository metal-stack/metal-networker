package netconf

import (
	"fmt"

	mn "github.com/metal-stack/metal-lib/pkg/net"
	"github.com/metal-stack/metal-networker/pkg/exec"
)

// ChronyServiceEnabler can enable chrony systemd service for the given VRF.
type ChronyServiceEnabler struct {
	VRF string
}

// NewChronyServiceEnabler constructs a new instance of this type.
func NewChronyServiceEnabler(kb KnowledgeBase) (ChronyServiceEnabler, error) {
	vrf, err := getDefaultRouteVRFName(kb)
	return ChronyServiceEnabler{VRF: vrf}, err
}

// Enable enables chrony systemd service for the given VRF to be started after boot.
func (c ChronyServiceEnabler) Enable() error {
	cmd := fmt.Sprintf("systemctl enable chrony@%s", c.VRF)
	log.Infof("running '%s' to enable chrony.'", cmd)

	return exec.NewVerboseCmd("bash", "-c", cmd).Run()
}

func getDefaultRouteVRFName(kb KnowledgeBase) (string, error) {
	networks := kb.GetNetworks(mn.External)
	for _, network := range networks {
		if containsDefaultRoute(network.Destinationprefixes) {
			vrf := fmt.Sprintf("vrf%d", *network.Vrf)
			return vrf, nil
		}
	}

	return "", fmt.Errorf("there is no network providing a default (0.0.0.0/0) route")
}

func containsDefaultRoute(prefixes []string) bool {
	for _, prefix := range prefixes {
		if prefix == IPv4ZeroCIDR || prefix == IPv6ZeroCIDR {
			return true
		}
	}
	return false
}
