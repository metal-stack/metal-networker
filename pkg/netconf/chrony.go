package netconf

import (
	"fmt"
	"os/exec"
)

// ChronyServiceEnabler can enable chrony systemd service for the given VRF.
type ChronyServiceEnabler struct {
	VRF string
}

// NewChronyServiceEnabler constructs a new instance of this type.
func NewChronyServiceEnabler(kb KnowledgeBase) (ChronyServiceEnabler, error) {
	vrf, err := kb.getDefaultRouteVRFName()
	return ChronyServiceEnabler{VRF: vrf}, err
}

// Enable enables chrony systemd service for the given VRF to be started after boot.
func (c ChronyServiceEnabler) Enable() error {
	cmd := fmt.Sprintf("systemctl enable chrony@%s", c.VRF)
	log.Infof("running '%s' to enable chrony.'", cmd)

	_, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	return err
}

func containsDefaultRoute(prefixes []string) bool {
	for _, prefix := range prefixes {
		if prefix == IPv4ZeroCIDR || prefix == IPv6ZeroCIDR {
			return true
		}
	}
	return false
}
