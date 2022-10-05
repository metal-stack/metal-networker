package cmd

import (
	"github.com/metal-stack/metal-networker/pkg/netconf"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/metal-stack/v"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

const (
	flagInputName = "input"
)

var (
	log *zap.SugaredLogger
	_   = initializeCmds()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "metal-networker",
	Short: "Configure network of bare metal servers",
	Long: `"metal-networker" is a self-sufficient tool to configure network related aspects of a bare metal server.
A bare metal server can be treated either as 'machine' or 'firewall'.`,
	Version: v.V.String(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	z, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	log = z.Sugar()

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
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
	rootCmd.PersistentFlags().StringP(flagInputName, "i", "", "Path to a YAML file containing network configuration")

	err := rootCmd.MarkPersistentFlagRequired(flagInputName)
	if err != nil {
		log.Warnf("error setting up cobra: %v", err)
	}

	return struct{}{}
}

// initConfig reads in ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}

// configure configures bare metal server depending on kind.
func configure(kind netconf.BareMetalType, cmd *cobra.Command) error {

	input, err := cmd.Flags().GetString(flagInputName)

	if err != nil {
		return err
	}

	log, err := newLogger(zapcore.InfoLevel)
	if err != nil {
		return err
	}
	log.Infof("running app version: %s", v.V.String())

	kb, err := netconf.New(log, input)
	if err != nil {
		return err
	}
	err = kb.Validate(kind)
	if err != nil {
		return err
	}

	c, err := netconf.NewConfigurator(kind, *kb)
	if err != nil {
		return err
	}
	c.Configure()
	log.Info("completed. Exiting..")

	return nil
}
func newLogger(level zapcore.Level) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zlog, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zlog.Sugar(), nil
}
