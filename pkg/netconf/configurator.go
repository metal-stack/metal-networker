package netconf

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/metal-stack/metal-networker/pkg/exec"
	"github.com/metal-stack/metal-networker/pkg/net"
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

	// CommonConfigurator contains information that is common to all configurators.
	CommonConfigurator struct {
		Kb KnowledgeBase
	}

	// MachineConfigurator is a configurator that configures a bare metal server as 'machine'.
	MachineConfigurator struct {
		CommonConfigurator
	}

	// FirewallConfigurator is a configurator that configures a bare metal server as 'firewall'.
	FirewallConfigurator struct {
		CommonConfigurator
		EnableDNSProxy bool
		EnableIDS      bool
	}
)

type unitConfiguration struct {
	unit             string
	templateFile     string
	constructApplier func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error)
	enabled          bool
}

// NewConfigurator creates a new configurator.
func NewConfigurator(kind BareMetalType, kb KnowledgeBase) Configurator {
	var result Configurator

	switch kind {
	case Firewall:
		fw := FirewallConfigurator{}
		fw.CommonConfigurator = CommonConfigurator{kb}
		result = fw
	case Machine:
		m := MachineConfigurator{}
		m.CommonConfigurator = CommonConfigurator{kb}
		result = m
	default:
		log.Fatalf("Unknown kind of configurator: %v", kind)
	}

	return result
}

// Configure applies configuration to a bare metal server to function as 'machine'.
func (configurator MachineConfigurator) Configure() {
	applyCommonConfiguration(Machine, configurator.Kb)
}

// Configure applies configuration to a bare metal server to function as 'firewall'.
func (configurator FirewallConfigurator) Configure() {
	kb := configurator.Kb
	applyCommonConfiguration(Firewall, kb)

	configurator.ConfigureNftables()

	chrony, err := NewChronyServiceEnabler(configurator.Kb)
	if err != nil {
		log.Warnf("failed to configure Chrony: %v", err)
	} else {
		err := chrony.Enable()
		if err != nil {
			log.Errorf("enabling Chrony failed: %v", err)
		}
	}

	for _, u := range configurator.getUnits() {
		src := mustTmpFile(u.unit)
		validatorService := ServiceValidator{src}
		nfe, err := u.constructApplier(configurator.Kb, validatorService)

		if err != nil {
			log.Warnf("failed to deploy %s service : %v", u.unit, err)
		}

		applyAndCleanUp(nfe, u.templateFile, src, path.Join(SystemdUnitPath, u.unit), FileModeSystemd)

		if u.enabled {
			mustEnableUnit(u.unit)
		}
	}

	configurator.ConfigureSuricataDefaults()
	configurator.ConfigureSuricata()
}

func (configurator FirewallConfigurator) ConfigureNftables() {
	src := mustTmpFile("nftrules_")
	validator := NftablesValidator{src}
	applier := NewNftablesConfigApplier(configurator.Kb, validator, configurator.EnableDNSProxy)
	applyAndCleanUp(applier, TplNftables, src, "/etc/nftables/rules", FileModeDefault)
}

func (configurator FirewallConfigurator) ConfigureSuricataDefaults() {
	src := mustTmpFile("suricata_")
	applier, err := NewSuricataDefaultsApplier(configurator.Kb, src)
	if err != nil {
		log.Warnf("failed to configure suricata defaults: %v", err)
	}
	applyAndCleanUp(applier, tplSuricataDefaults, src, "/etc/default/suricata", FileModeSixFourFour)
}

func (configurator FirewallConfigurator) ConfigureSuricata() {
	src := mustTmpFile("suricata.yaml_")
	applier, err := NewSuricataConfigApplier(configurator.Kb, src, configurator.EnableIDS)
	if err != nil {
		log.Warnf("failed to configure suricata: %v", err)
	}
	applyAndCleanUp(applier, TplSuricataConfig, src, "/etc/suricata/suricata.yaml", FileModeSixFourFour)
}

func (configurator FirewallConfigurator) getUnits() []unitConfiguration {
	return []unitConfiguration{
		{
			unit:         systemdUnitDroptailer,
			templateFile: tplDroptailer,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewDroptailerServiceApplier(kb, v)
			},
			enabled: false, // will be enabled in the case of k8s deployments with ignition on first boot
		},
		{
			unit:         systemdUnitFirewallController,
			templateFile: tplFirewallController,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewFirewallControllerServiceApplier(kb, v)
			},
			enabled: false, // will be enabled in the case of k8s deployments with ignition on first boot
		},
		{
			unit:         systemdUnitNftablesExporter,
			templateFile: tplNftablesExporter,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewNftablesExporterServiceApplier(kb, v)
			},
			enabled: true,
		},
		{
			unit:         systemdUnitNodeExporter,
			templateFile: tplNodeExporter,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewNodeExporterServiceApplier(kb, v)
			},
			enabled: true,
		},
		{
			unit:         systemdUnitSuricataUpdate,
			templateFile: tplSuricataUpdate,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewSuricataUpdateServiceApplier(kb, v)
			},
			enabled: true,
		},
	}
}

func applyCommonConfiguration(kind BareMetalType, kb KnowledgeBase) {
	a := NewIfacesApplier(kind, kb)
	a.Apply()

	src := mustTmpFile("hosts_")
	applier := NewHostsApplier(kb, src)
	applyAndCleanUp(applier, TplHosts, src, "/etc/hosts", FileModeDefault)

	src = mustTmpFile("hostname_")
	applier = NewHostnameApplier(kb, src)
	applyAndCleanUp(applier, TplHostname, src, "/etc/hostname", FileModeSixFourFour)

	src = mustTmpFile("frr_")
	applier = NewFrrConfigApplier(kind, kb, src)
	tpl := TplFirewallFRR

	if kind == Machine {
		tpl = TplMachineFRR
	}

	applyAndCleanUp(applier, tpl, src, "/etc/frr/frr.conf", FileModeDefault)
}

func applyAndCleanUp(applier net.Applier, tpl, src, dest string, mode os.FileMode) {
	log.Infof("rendering %s to %s (mode: %s)", tpl, dest, mode)
	file := mustReadTpl(tpl)
	mustApply(applier, file, src, dest)

	err := os.Chmod(dest, mode)
	if err != nil {
		log.Errorf("error to chmod %s to %s", dest, mode)
	}

	_ = os.Remove(src)
}

func mustEnableUnit(unit string) {
	cmd := fmt.Sprintf("systemctl enable %s", unit)
	log.Infof("running '%s' to enable unit.'", cmd)

	err := exec.NewVerboseCmd("bash", "-c", cmd).Run()

	if err != nil {
		log.Panic(err)
	}
}

func mustApply(applier net.Applier, tpl, src, dest string) {
	t := template.Must(template.New(src).Parse(tpl))
	_, err := applier.Apply(*t, src, dest, false)

	if err != nil {
		log.Panic(err)
	}
}

func mustTmpFile(prefix string) string {
	f, err := os.CreateTemp(tmpPath, prefix)
	if err != nil {
		log.Panic(err)
	}

	err = f.Close()
	if err != nil {
		log.Panic(err)
	}

	return f.Name()
}
