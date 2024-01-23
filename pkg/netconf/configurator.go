package netconf

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"text/template"

	"github.com/metal-stack/metal-networker/pkg/exec"
	"github.com/metal-stack/metal-networker/pkg/net"
)

// BareMetalType defines the type of configuration to apply.
type BareMetalType int

const (
	// Firewall defines the bare metal server to function as firewall.
	Firewall BareMetalType = iota
	// Machine defines the bare metal server to function as machine.
	Machine
)
const (
	// fileModeSystemd represents a file mode that allows systemd to read e.g. /etc/systemd/network files.
	fileModeSystemd = 0644
	// fileModeSixFourFour represents file mode 0644
	fileModeSixFourFour = 0644
	// fileModeDefault represents the default file mode sufficient e.g. to /etc/network/interfaces or /etc/frr.conf.
	fileModeDefault = 0600
	// systemdUnitPath is the path where systemd units will be generated.
	systemdUnitPath = "/etc/systemd/system/"
)

var (
	// systemdNetworkPath is the path where systemd-networkd expects its configuration files.
	systemdNetworkPath = "/etc/systemd/network"
	// tmpPath is the path where temporary files are stored for validation before they are moved to their intended place.
	tmpPath = "/etc/metal/networker/"
)

// ForwardPolicy defines how packets in the forwarding chain are handled, can be either drop or accept.
// drop will be the standard for firewalls which are not managed by kubernetes resources (CWNPs)
type ForwardPolicy string

const (
	// ForwardPolicyDrop drops packets which try to go through the forwarding chain
	ForwardPolicyDrop = ForwardPolicy("drop")
	// ForwardPolicyAccept accepts packets which try to go through the forwarding chain
	ForwardPolicyAccept = ForwardPolicy("accept")
)

type (
	// Configurator is an interface to configure bare metal servers.
	Configurator interface {
		Configure(forwardPolicy ForwardPolicy)
		ConfigureNftables(forwardPolicy ForwardPolicy)
	}

	// machineConfigurator is a configurator that configures a bare metal server as 'machine'.
	machineConfigurator struct {
		c config
	}

	// firewallConfigurator is a configurator that configures a bare metal server as 'firewall'.
	firewallConfigurator struct {
		c              config
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
func NewConfigurator(kind BareMetalType, c config, enableDNS bool) (Configurator, error) {
	switch kind {
	case Firewall:
		return firewallConfigurator{
			c:              c,
			enableDNSProxy: enableDNS,
		}, nil
	case Machine:
		return machineConfigurator{
			c: c,
		}, nil
	default:
		return nil, fmt.Errorf("unknown type:%d", kind)
	}
}

// Configure applies configuration to a bare metal server to function as 'machine'.
func (mc machineConfigurator) Configure(forwardPolicy ForwardPolicy) {
	applyCommonConfiguration(mc.c.log, Machine, mc.c)
}

// ConfigureNftables is empty function that exists just to satisfy the Configurator interface
func (mc machineConfigurator) ConfigureNftables(forwardPolicy ForwardPolicy) {}

// Configure applies configuration to a bare metal server to function as 'firewall'.
func (fc firewallConfigurator) Configure(forwardPolicy ForwardPolicy) {
	kb := fc.c
	applyCommonConfiguration(fc.c.log, Firewall, kb)

	fc.ConfigureNftables(forwardPolicy)

	chrony, err := newChronyServiceEnabler(fc.c)
	if err != nil {
		fc.c.log.Warn("failed to configure chrony", "error", err)
	} else {
		err := chrony.Enable()
		if err != nil {
			fc.c.log.Error("enabling chrony failed", "error", err)
		}
	}

	for _, u := range fc.getUnits() {
		src := mustTmpFile(u.unit)
		validatorService := serviceValidator{src}
		nfe, err := u.constructApplier(fc.c, validatorService)

		if err != nil {
			fc.c.log.Warn("failed to deploy", "unit", u.unit, "error", err)
		}

		applyAndCleanUp(fc.c.log, nfe, u.templateFile, src, path.Join(systemdUnitPath, u.unit), fileModeSystemd, false)

		if u.enabled {
			mustEnableUnit(fc.c.log, u.unit)
		}
	}

	src := mustTmpFile("suricata_")
	applier, err := newSuricataDefaultsApplier(kb, src)

	if err != nil {
		fc.c.log.Warn("failed to configure suricata defaults", "error", err)
	}

	applyAndCleanUp(fc.c.log, applier, tplSuricataDefaults, src, "/etc/default/suricata", fileModeSixFourFour, false)

	src = mustTmpFile("suricata.yaml_")
	applier, err = newSuricataConfigApplier(kb, src)

	if err != nil {
		fc.c.log.Warn("failed to configure suricata", "error", err)
	}

	applyAndCleanUp(fc.c.log, applier, tplSuricataConfig, src, "/etc/suricata/suricata.yaml", fileModeSixFourFour, false)
}

func (fc firewallConfigurator) ConfigureNftables(forwardPolicy ForwardPolicy) {
	src := mustTmpFile("nftrules_")
	validator := NftablesValidator{
		path: src,
		log:  fc.c.log,
	}
	applier := newNftablesConfigApplier(fc.c, validator, fc.enableDNSProxy, forwardPolicy)
	applyAndCleanUp(fc.c.log, applier, TplNftables, src, "/etc/nftables/rules", fileModeDefault, true)
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

func applyCommonConfiguration(log *slog.Logger, kind BareMetalType, kb config) {
	a := newIfacesApplier(kind, kb)
	a.Apply()

	src := mustTmpFile("hosts_")
	applier := newHostsApplier(kb, src)
	applyAndCleanUp(log, applier, tplHosts, src, "/etc/hosts", fileModeDefault, false)

	src = mustTmpFile("hostname_")
	applier = newHostnameApplier(kb, src)
	applyAndCleanUp(log, applier, tplHostname, src, "/etc/hostname", fileModeSixFourFour, false)

	src = mustTmpFile("frr_")
	applier = NewFrrConfigApplier(kind, kb, src)
	tpl := TplFirewallFRR

	if kind == Machine {
		tpl = TplMachineFRR
	}

	applyAndCleanUp(log, applier, tpl, src, "/etc/frr/frr.conf", fileModeDefault, false)
}

func applyAndCleanUp(log *slog.Logger, applier net.Applier, tpl, src, dest string, mode os.FileMode, reload bool) {
	log.Info("rendering", "template", tpl, "destination", dest, "mode", mode)
	file := mustReadTpl(tpl)
	mustApply(applier, file, src, dest, reload)

	err := os.Chmod(dest, mode)
	if err != nil {
		log.Error("unable change mode", "file", dest, "mode", mode, "error", err)
	}

	_ = os.Remove(src)
}

func mustEnableUnit(log *slog.Logger, unit string) {
	cmd := fmt.Sprintf("systemctl enable %s", unit)
	log.Info("enable unit", "command", cmd)

	err := exec.NewVerboseCmd("bash", "-c", cmd).Run()

	if err != nil {
		panic(err)
	}
}

func mustApply(applier net.Applier, tpl, src, dest string, reload bool) {
	t := template.Must(template.New(src).Parse(tpl))
	_, err := applier.Apply(*t, src, dest, reload)

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
