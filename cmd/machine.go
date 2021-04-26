package cmd

import (
	"github.com/metal-stack/metal-networker/pkg/netconf"
	"github.com/spf13/cobra"
)

// machineCmd represents the machine command
var (
	machineCmd = &cobra.Command{
		Use:     "machine",
		Aliases: []string{"m"},
		Short:   "Treat a bare metal server as a 'machine'",
		Long:    `"metal-networker machine" treats a bare metal server as 'machine'`,
	}
	machineConfigureCmd = &cobra.Command{
		Use:     "configure",
		Aliases: []string{"c"},
		Short:   "Configures network aspects",
		Long: `"metal-networker machine configure" configures network related aspects of a bare metal server
to function as a 'machine'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return configure(netconf.Machine, cmd)
		},
	}
)
