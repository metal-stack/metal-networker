package cmd

import (
	"git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf"

	"github.com/spf13/cobra"
)

// firewallCmd represents the firewall command
var (
	firewallCmd = &cobra.Command{
		Use:     "firewall",
		Aliases: []string{"fw"},
		Short:   "Treat a bare metal server as a 'firewall'",
		Long:    `"metal-networker firewall" treats a bare metal server as 'firewall'`,
	}
	firewallConfigureCmd = &cobra.Command{
		Use:     "configure",
		Aliases: []string{"c"},
		Short:   "Configures network aspects",
		Long: `"metal-networker firewall configure" configures network aspects of a bare metal server to function as
a 'firewall'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return configure(netconf.Firewall, cmd)
		},
	}
)
