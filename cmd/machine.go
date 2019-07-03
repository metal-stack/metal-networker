package cmd

import (
	"git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf"
	"github.com/spf13/cobra"
)

// FlagInputName defines the name of the flag to YAML input.
const FlagInputName = "input"

// machineCmd represents the machine command
var (
	machineCmd = &cobra.Command{
		Use:   "machine",
		Short: "Treat a bare metal server as a 'machine'",
		Long:  `"metal-networker machine" treats a bare metal server as 'machine'`,
	}
	machineConfigureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Configures network aspects",
		Long: `"metal-networker machine configure" configures network related aspects of a bare metal server
to function as a 'machine'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Configure(netconf.Machine, cmd, args)
		},
	}
)
