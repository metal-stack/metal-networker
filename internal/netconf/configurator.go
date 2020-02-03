package netconf

import (
	"fmt"
	"io/ioutil"
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
	// SystemdUnitPath is the path where systemd units will be generated,
	SystemdUnitPath = "/etc/systemd/system/"
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
		log *zap.SugaredLogger
	}

	// FirewallConfigurator is a configurator that configures a bare metal server as 'firewall'.
	FirewallConfigurator struct {
		CommonConfigurator
		log *zap.SugaredLogger
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
		fw.log = kb.log
		result = fw
	case Machine:
		m := MachineConfigurator{}
		m.CommonConfigurator = CommonConfigurator{kb}
		m.log = kb.log
		result = m
	default:
		kb.log.Fatalf("Unknown kind of configurator: %v", kind)
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

	src := mustTmpFile("rules.v4_", kb.log)
	validatorIPv4 := NftablesV4Validator{NftablesValidator{src, kb.log}}
	applier := NewNftablesConfigApplier(configurator.Kb, validatorIPv4)
	kb.applyAndCleanUp(applier, TplNftablesV4, src, "/etc/nftables/rules.v4", FileModeDefault)
	src = mustTmpFile("rules.v6_", kb.log)
	validatorIPv6 := NftablesV6Validator{NftablesValidator{src, kb.log}}
	applier = NewNftablesConfigApplier(configurator.Kb, validatorIPv6)
	kb.applyAndCleanUp(applier, TplNftablesV6, src, "/etc/nftables/rules.v6", FileModeDefault)

	chrony, err := NewChronyServiceEnabler(configurator.Kb)
	if err != nil {
		kb.log.Warnf("failed to configure Chrony: %v", err)
	} else {
		err := chrony.Enable()
		if err != nil {
			kb.log.Errorf("enabling Chrony failed: %v", err)
		}
	}

	for _, u := range configurator.getUnits() {
		src = mustTmpFile(u.unit, kb.log)
		validatorService := ServiceValidator{src}
		nfe, err := u.constructApplier(configurator.Kb, validatorService)

		if err != nil {
			kb.log.Warnf("failed to deploy %s service : %v", u.unit, err)
		}

		kb.applyAndCleanUp(nfe, u.templateFile, src, path.Join(SystemdUnitPath, u.unit), FileModeSystemd)

		if u.enabled {
			mustEnableUnit(u.unit, kb.log)
		}
	}
}

func (configurator FirewallConfigurator) getUnits() []unitConfiguration {
	return []unitConfiguration{
		{
			unit:         SystemdUnitDroptailer,
			templateFile: TplDroptailer,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewDroptailerServiceApplier(kb, v)
			},
			enabled: false, // will be enabled in the case of k8s deployments with ignition on first boot
		},
		{
			unit:         SystemdUnitFirewallPolicyController,
			templateFile: TplFirewallPolicyController,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewFirewallPolicyControllerServiceApplier(kb, v)
			},
			enabled: false, // will be enabled in the case of k8s deployments with ignition on first boot
		},
		{
			unit:         SystemdUnitNftablesExporter,
			templateFile: TplNftablesExporter,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewNftablesExporterServiceApplier(kb, v)
			},
			enabled: true,
		},
		{
			unit:         SystemdUnitNodeExporter,
			templateFile: TplNodeExporter,
			constructApplier: func(kb KnowledgeBase, v ServiceValidator) (net.Applier, error) {
				return NewNodeExporterServiceApplier(kb, v)
			},
			enabled: true,
		},
	}
}

func applyCommonConfiguration(kind BareMetalType, kb KnowledgeBase) {
	src := mustTmpFile("interfaces_", kb.log)
	applier := NewIfacesConfigApplier(kind, kb, src)
	tpl := TplFirewallIfaces

	if kind == Machine {
		tpl = TplMachineIfaces
	}

	kb.applyAndCleanUp(applier, tpl, src, "/etc/network/interfaces", FileModeDefault)

	src = mustTmpFile("hosts_", kb.log)
	applier = NewHostsApplier(kb, src)
	kb.applyAndCleanUp(applier, TplHosts, src, "/etc/hosts", FileModeDefault)

	src = mustTmpFile("hostname_", kb.log)
	applier = NewHostnameApplier(kb, src)
	kb.applyAndCleanUp(applier, TplHostname, src, "/etc/hostname", FileModeSixFourFour)

	src = mustTmpFile("frr_", kb.log)
	applier = NewFrrConfigApplier(kind, kb, src)
	tpl = TplFirewallFRR

	if kind == Machine {
		tpl = TplMachineFRR
	}

	kb.applyAndCleanUp(applier, tpl, src, "/etc/frr/frr.conf", FileModeDefault)

	offset := 1

	for i, nic := range kb.Nics {
		prefix := fmt.Sprintf("lan%d_link_", i)
		src = mustTmpFile(prefix, kb.log)
		applier = NewSystemdLinkApplier(kind, kb.Machineuuid, i, nic, src, kb.log)
		dest := fmt.Sprintf("/etc/systemd/network/%d0-lan%d.link", i+offset, i)
		kb.applyAndCleanUp(applier, TplSystemdLink, src, dest, FileModeSystemd)

		prefix = fmt.Sprintf("lan%d_network_", i)
		src = mustTmpFile(prefix, kb.log)
		applier = NewSystemdNetworkApplier(kb.Machineuuid, i, src, kb.log)
		dest = fmt.Sprintf("/etc/systemd/network/%d0-lan%d.network", i+offset, i)
		kb.applyAndCleanUp(applier, TplSystemdNetwork, src, dest, FileModeSystemd)
	}
}

func (kb KnowledgeBase) applyAndCleanUp(applier net.Applier, tpl, src, dest string, mode os.FileMode) {
	kb.log.Infof("rendering %s to %s (mode: %ui)", tpl, dest, mode)
	file := mustRead(tpl, kb.log)
	mustApply(applier, file, src, dest, kb.log)

	err := os.Chmod(dest, mode)
	if err != nil {
		kb.log.Errorf("error to chmod %s to %ui", dest, mode)
	}

	_ = os.Remove(src)
}

func mustEnableUnit(unit string, log *zap.SugaredLogger) {
	cmd := fmt.Sprintf("systemctl enable %s", unit)
	log.Infof("running '%s' to enable unit.'", cmd)

	err := exec.NewVerboseCmd("bash", "-c", cmd).Run()

	if err != nil {
		log.Panic(err)
	}
}

func mustApply(applier net.Applier, tpl, src, dest string, log *zap.SugaredLogger) {
	t := template.Must(template.New(TplFirewallIfaces).Parse(tpl))
	err := applier.Apply(*t, src, dest, false)

	if err != nil {
		log.Panic(err)
	}
}

func mustRead(name string, log *zap.SugaredLogger) string {
	c, err := ioutil.ReadFile(name)
	if err != nil {
		log.Panic(err)
	}

	return string(c)
}

func mustTmpFile(prefix string, log *zap.SugaredLogger) string {
	f, err := ioutil.TempFile("/etc/metal/networker/", prefix)
	if err != nil {
		log.Panic(err)
	}

	err = f.Close()
	if err != nil {
		log.Panic(err)
	}

	return f.Name()
}
