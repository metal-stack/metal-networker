package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/exec"
	"go.uber.org/zap"
)

// ChronyServiceEnabler can enable chrony systemd service for the given VRF.
type ChronyServiceEnabler struct {
	vrf string
	log *zap.SugaredLogger
}

// NewChronyServiceEnabler constructs a new instance of this type.
func NewChronyServiceEnabler(kb config) (ChronyServiceEnabler, error) {
	vrf, err := kb.getDefaultRouteVRFName()
	return ChronyServiceEnabler{
		vrf: vrf,
		log: kb.log,
	}, err
}

// Enable enables chrony systemd service for the given VRF to be started after boot.
func (c ChronyServiceEnabler) Enable() error {
	cmd := fmt.Sprintf("systemctl enable chrony@%s", c.vrf)
	c.log.Infof("running '%s' to enable chrony.'", cmd)

	return exec.NewVerboseCmd("bash", "-c", cmd).Run()
}

func containsDefaultRoute(prefixes []string) bool {
	for _, prefix := range prefixes {
		if prefix == IPv4ZeroCIDR || prefix == IPv6ZeroCIDR {
			return true
		}
	}
	return false
}
