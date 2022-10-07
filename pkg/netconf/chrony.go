package netconf

import (
	"fmt"

	"github.com/metal-stack/metal-networker/pkg/exec"
	"go.uber.org/zap"
)

// chronyServiceEnabler can enable chrony systemd service for the given VRF.
type chronyServiceEnabler struct {
	vrf string
	log *zap.SugaredLogger
}

// newChronyServiceEnabler constructs a new instance of this type.
func newChronyServiceEnabler(kb config) (chronyServiceEnabler, error) {
	vrf, err := kb.getDefaultRouteVRFName()
	return chronyServiceEnabler{
		vrf: vrf,
		log: kb.log,
	}, err
}

// Enable enables chrony systemd service for the given VRF to be started after boot.
func (c chronyServiceEnabler) Enable() error {
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
