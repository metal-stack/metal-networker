package netconf

import (
	"fmt"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/pkg/exec"
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
	networks := kb.GetNetworks(External)
	for _, n := range networks {
		for _, d := range n.Destinationprefixes {
			if d == "0.0.0.0/0" {
				vrf := fmt.Sprintf("vrf%d", n.Vrf)
				return vrf, nil
			}
		}
	}
	return "", fmt.Errorf("there is no network providing a default (0.0.0.0/0) route")
}
