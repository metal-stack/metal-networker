package netconf

import (
	"fmt"
	"log/slog"

	"github.com/metal-stack/metal-networker/pkg/exec"
)

// chronyServiceEnabler can enable chrony systemd service for the given VRF.
type chronyServiceEnabler struct {
	vrf string
	log *slog.Logger
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
	c.log.Info("enable chrony", "command", cmd)

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
