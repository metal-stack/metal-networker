package cmd

import (
	"fmt"
	"os"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf"

	"git.f-i-ts.de/cloud-native/metallib/zapup"

	"github.com/metal-pod/v"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	log     = zapup.MustRootLogger().Sugar()
	_       = initializeCmds()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "metal-networker",
	Short: "Configure network of bare metal servers",
	Long: `"metal-networker" is a self-sufficient tool to configure network related aspects of a bare metal server.
A bare metal server can be treated either as 'machine' or 'firewall'.`,
	Version: v.V.String(),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// One must not use init() functions in go.
// As a workaround initializeCmds function is used.
// See https://medium.com/random-go-tips/init-without-init-ebf2f62e7c4a
func initializeCmds() struct{} {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.AddCommand(machineCmd)
	rootCmd.AddCommand(firewallCmd)

	machineCmd.AddCommand(machineConfigureCmd)
	firewallCmd.AddCommand(firewallConfigureCmd)

	// Here you will define your flags and configuration settings.
	//rootCmd.MarkPersistentFlagRequired(FlagInputName)
	//rootCmd.PersistentFlags().StringP(FlagInputName, "i", "", "Path to a YAML file containing network configuration")
	rootCmd.PersistentFlags().StringP(FlagInputName, "i", "", "Path to a YAML file containing network configuration")
	err := rootCmd.MarkPersistentFlagRequired(FlagInputName)
	if err != nil {
		log.Warnf("error setting up cobra: %v", err)
	}

	return struct{}{}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".metal-networker" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".metal-networker")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Configure configures bare metal server depending on kind.
func Configure(kind netconf.BareMetalType, cmd *cobra.Command, args []string) error {

	log.Infof("running app version: %s", v.V.String())
	input, err := cmd.Flags().GetString(FlagInputName)
	if err != nil {
		return err
	}

	kb := netconf.NewKnowledgeBase(input)
	netconf.NewConfigurator(kind, kb).Configure()
	log.Info("completed. Exiting..")

	return nil
}
