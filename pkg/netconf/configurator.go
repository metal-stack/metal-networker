package netconf

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/metal-stack/metal-networker/pkg/exec"
	"github.com/metal-stack/metal-networker/pkg/net"
	"go.uber.org/zap"
)

// BareMetalType defines the type of configuration to apply.
type BareMetalType int

const (
	// FileModeSystemd represents a file mode that allows systemd to read e.g. /etc/systemd/network files.
	FileModeSystemd = 0644
	// FileModeSixFourFour represents file mode 0644
	FileModeSixFourFour = 0644
	// FileModeDefault represents the default file mode sufficient e.g. to /etc/network/interfaces or /etc/frr.conf.
	FileModeDefault = 0600
	// Firewall defines the bare metal server to function as firewall.
	Firewall BareMetalType = iota
	// Machine defines the bare metal server to function as machine.
	Machine
	// SystemdUnitPath is the path where systemd units will be generated.
	SystemdUnitPath = "/etc/systemd/system/"
)

var (
	// systemdNetworkPath is the path where systemd-networkd expects its configuration files.
	systemdNetworkPath = "/etc/systemd/network"
	// tmpPath is the path where temporary files are stored for validation before they are moved to their intended place.
	tmpPath = "/etc/metal/networker/"
)

type (
	// Configurator is an interface to configure bare metal servers.
	Configurator interface {
		Configure()
	}

	// commonConfigurator contains information that is common to all configurators.
	commonConfigurator struct {
		c config
	}

	// machineConfigurator is a configurator that configures a bare metal server as 'machine'.
	machineConfigurator struct {
		commonConfigurator
	}

	// firewallConfigurator is a configurator that configures a bare metal server as 'firewall'.
	firewallConfigurator struct {
		commonConfigurator
		enableDNSProxy bool
	}
)

type unitConfiguration struct {
	unit             string
	templateFile     string
	constructApplier func(kb config, v serviceValidator) (net.Applier, error)
	enabled          bool
}

// NewConfigurator creates a new configurator.
func NewConfigurator(kind BareMetalType, c config) Configurator {
	var result Configurator

	switch kind {
	case Firewall:
		fw := firewallConfigurator{}
		fw.commonConfigurator = commonConfigurator{c}
		result = fw
	case Machine:
		m := machineConfigurator{}
		m.commonConfigurator = commonConfigurator{c}
		result = m
	default:
		c.log.Fatalf("Unknown kind of configurator: %v", kind)
	}

	return result
}

// Configure applies configuration to a bare metal server to function as 'machine'.
func (mc machineConfigurator) Configure() {
	applyCommonConfiguration(mc.c.log, Machine, mc.c)
}

// Configure applies configuration to a bare metal server to function as 'firewall'.
func (fc firewallConfigurator) Configure() {
	kb := fc.c
	applyCommonConfiguration(fc.c.log, Firewall, kb)

	fc.configureNftables()

	chrony, err := newChronyServiceEnabler(fc.c)
	if err != nil {
		fc.c.log.Warnf("failed to configure Chrony: %v", err)
	} else {
		err := chrony.Enable()
		if err != nil {
			fc.c.log.Errorf("enabling Chrony failed: %v", err)
		}
	}

	for _, u := range fc.getUnits() {
		src := mustTmpFile(u.unit)
		validatorService := serviceValidator{src}
		nfe, err := u.constructApplier(fc.c, validatorService)

		if err != nil {
			fc.c.log.Warnf("failed to deploy %s service : %v", u.unit, err)
		}

		applyAndCleanUp(fc.c.log, nfe, u.templateFile, src, path.Join(SystemdUnitPath, u.unit), FileModeSystemd)

		if u.enabled {
			mustEnableUnit(fc.c.log, u.unit)
		}
	}

	src := mustTmpFile("suricata_")
	applier, err := newSuricataDefaultsApplier(kb, src)

	if err != nil {
		fc.c.log.Warnf("failed to configure suricata defaults: %v", err)
	}

	applyAndCleanUp(fc.c.log, applier, tplSuricataDefaults, src, "/etc/default/suricata", FileModeSixFourFour)

	src = mustTmpFile("suricata.yaml_")
	applier, err = newSuricataConfigApplier(kb, src)

	if err != nil {
		fc.c.log.Warnf("failed to configure suricata: %v", err)
	}

	applyAndCleanUp(fc.c.log, applier, tplSuricataConfig, src, "/etc/suricata/suricata.yaml", FileModeSixFourFour)
}

func (fc firewallConfigurator) configureNftables() {
	src := mustTmpFile("nftrules_")
	validator := NftablesValidator{
		path: src,
		log:  fc.c.log,
	}
	applier := newNftablesConfigApplier(fc.c, validator, fc.enableDNSProxy)
	applyAndCleanUp(fc.c.log, applier, TplNftables, src, "/etc/nftables/rules", FileModeDefault)
}

func (fc firewallConfigurator) getUnits() (units []unitConfiguration) {
	units = []unitConfiguration{
		{
			unit:         systemdUnitDroptailer,
			templateFile: tplDroptailer,
			constructApplier: func(kb config, v serviceValidator) (net.Applier, error) {
				return newDroptailerServiceApplier(kb, v)
			},
			enabled: false, // will be enabled in the case of k8s deployments with ignition on first boot
		},
		{
			unit:         systemdUnitFirewallController,
			templateFile: tplFirewallController,
			constructApplier: func(kb config, v serviceValidator) (net.Applier, error) {
				return newFirewallControllerServiceApplier(kb, v)
			},
			enabled: false, // will be enabled in the case of k8s deployments with ignition on first boot
		},
		{
			unit:         systemdUnitNftablesExporter,
			templateFile: tplNftablesExporter,
			constructApplier: func(kb config, v serviceValidator) (net.Applier, error) {
				return NewNftablesExporterServiceApplier(kb, v)
			},
			enabled: true,
		},
		{
			unit:         systemdUnitNodeExporter,
			templateFile: tplNodeExporter,
			constructApplier: func(kb config, v serviceValidator) (net.Applier, error) {
				return newNodeExporterServiceApplier(kb, v)
			},
			enabled: true,
		},
		{
			unit:         systemdUnitSuricataUpdate,
			templateFile: tplSuricataUpdate,
			constructApplier: func(kb config, v serviceValidator) (net.Applier, error) {
				return newSuricataUpdateServiceApplier(kb, v)
			},
			enabled: true,
		},
	}

	if fc.c.VPN != nil {
		units = append(units, unitConfiguration{
			unit:         systemdUnitTailscaled,
			templateFile: tplTailscaled,
			constructApplier: func(kb config, v serviceValidator) (net.Applier, error) {
				return newTailscaledServiceApplier(kb, v)
			},
			enabled: true,
		}, unitConfiguration{
			unit:         systemdUnitTailscale,
			templateFile: tplTailscale,
			constructApplier: func(kb config, v serviceValidator) (net.Applier, error) {
				return newTailscaleServiceApplier(kb, v)
			},
			enabled: true,
		})
	}

	return units
}

func applyCommonConfiguration(log *zap.SugaredLogger, kind BareMetalType, kb config) {
	a := newIfacesApplier(kind, kb)
	a.Apply()

	src := mustTmpFile("hosts_")
	applier := newHostsApplier(kb, src)
	applyAndCleanUp(log, applier, tplHosts, src, "/etc/hosts", FileModeDefault)

	src = mustTmpFile("hostname_")
	applier = newHostnameApplier(kb, src)
	applyAndCleanUp(log, applier, tplHostname, src, "/etc/hostname", FileModeSixFourFour)

	src = mustTmpFile("frr_")
	applier = NewFrrConfigApplier(kind, kb, src)
	tpl := TplFirewallFRR

	if kind == Machine {
		tpl = TplMachineFRR
	}

	applyAndCleanUp(log, applier, tpl, src, "/etc/frr/frr.conf", FileModeDefault)
}

func applyAndCleanUp(log *zap.SugaredLogger, applier net.Applier, tpl, src, dest string, mode os.FileMode) {
	log.Infof("rendering %s to %s (mode: %s)", tpl, dest, mode)
	file := mustReadTpl(tpl)
	mustApply(applier, file, src, dest)

	err := os.Chmod(dest, mode)
	if err != nil {
		log.Errorf("error to chmod %s to %s", dest, mode)
	}

	_ = os.Remove(src)
}

func mustEnableUnit(log *zap.SugaredLogger, unit string) {
	cmd := fmt.Sprintf("systemctl enable %s", unit)
	log.Infof("running '%s' to enable unit.'", cmd)

	err := exec.NewVerboseCmd("bash", "-c", cmd).Run()

	if err != nil {
		panic(err)
	}
}

func mustApply(applier net.Applier, tpl, src, dest string) {
	t := template.Must(template.New(src).Parse(tpl))
	_, err := applier.Apply(*t, src, dest, false)

	if err != nil {
		panic(err)
	}
}

func mustTmpFile(prefix string) string {
	f, err := os.CreateTemp(tmpPath, prefix)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	return f.Name()
}
