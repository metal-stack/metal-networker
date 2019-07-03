package netconf

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

// BareMetalType defines the type of configuration to apply.
type BareMetalType int

const (
	// Firewall defines the bare metal server to function as firewall.
	Firewall BareMetalType = iota
	// Machine defines the bare metal server to function as machine.
	Machine
)

// Configurator is an interface to configure bare metal servers.
type Configurator interface {
	Configure()
}

// CommonConfigurator contains information that is common to all configurators.
type CommonConfigurator struct {
	Kb KnowledgeBase
}

// MachineConfigurator is a configurator that configures a bare metal server as 'machine'.
type MachineConfigurator struct {
	*CommonConfigurator
}

// FirewallConfigurator is a configurator that configures a bare metal server as 'firewall'.
type FirewallConfigurator struct {
	*CommonConfigurator
}

// NewConfigurator creates a new configurator.
func NewConfigurator(kind BareMetalType, kb KnowledgeBase) Configurator {
	var result Configurator
	switch kind {
	case Firewall:
		fw := FirewallConfigurator{}
		fw.CommonConfigurator = &CommonConfigurator{kb}
		result = fw
	case Machine:
		m := MachineConfigurator{}
		m.CommonConfigurator = &CommonConfigurator{kb}
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
	applyCommonConfiguration(Firewall, configurator.Kb)

	src := mustTmpFile("rules.v4_")
	applier := NewIptablesConfigApplier(configurator.Kb, src)
	applyAndCleanUp(applier, TplIptables, src, "/etc/iptables/rules.v4")

	chrony, err := NewChronyServiceEnabler(configurator.Kb)
	if err != nil {
		log.Warnf("failed to configure Chrony: %v", err)
	} else {
		err := chrony.Enable()
		if err != nil {
			log.Errorf("enabling Chrony failed: %v", err)
		}
	}
}

func applyCommonConfiguration(kind BareMetalType, kb KnowledgeBase) {
	src := mustTmpFile("interfaces_")
	applier := NewIfacesConfigApplier(kind, kb, src)
	applyAndCleanUp(applier, TplFirewallIfaces, src, "/etc/network/interfaces")

	src = mustTmpFile("hosts_")
	applier = NewHostsApplier(kb, src)
	applyAndCleanUp(applier, TplHosts, src, "/etc/hosts")

	src = mustTmpFile("frr_")
	applier = NewFrrConfigApplier(kind, kb, src)
	applyAndCleanUp(applier, TplFirewallFRR, src, "/etc/frr/frr.conf")

	for i, nic := range kb.Nics {
		prefix := fmt.Sprintf("lan%d_link_", i)
		src = mustTmpFile(prefix)
		applier = NewSystemdLinkApplier(kind, kb.Machineuuid, i, nic, src)
		dest := fmt.Sprintf("/etc/systemd/network/%d0-lan%d.link", i+1, i)
		applyAndCleanUp(applier, TplSystemdLink, src, dest)

		prefix = fmt.Sprintf("lan%d_network_", i)
		src = mustTmpFile(prefix)
		applier = NewSystemdNetworkApplier(kb.Machineuuid, i, src)
		dest = fmt.Sprintf("/etc/systemd/network/%d0-lan%d.network", i+1, i)
		applyAndCleanUp(applier, TplSystemdNetwork, src, dest)
	}
}

func applyAndCleanUp(applier network.Applier, tpl, src, dest string) {
	file := mustRead(tpl)
	mustApply(applier, file, src, dest)
	_ = os.Remove(src)
}

func mustApply(applier network.Applier, tpl, src, dest string) {
	t := template.Must(template.New(TplFirewallIfaces).Parse(tpl))
	log.Infof("applying changes to: %s", dest)
	err := applier.Apply(*t, src, dest, false)
	if err != nil {
		log.Panic(err)
	}
}

func mustRead(name string) string {
	log.Infof("reading template: %s", name)
	c, err := ioutil.ReadFile(name)
	if err != nil {
		log.Panic(err)
	}
	return string(c)
}

func mustTmpFile(prefix string) string {
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
