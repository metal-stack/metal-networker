package main

import (
	"io/ioutil"
	"os"
	"text/template"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/internal/system"

	"git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf"

	"github.com/metal-pod/v"

	"git.f-i-ts.de/cloud-native/metallib/zapup"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

var log = zapup.MustRootLogger().Sugar()

func main() {
	log.Infof("running app version: %s", v.V.String())

	a := mustArg(1)
	log.Infof("loading: %s", a)
	d := netconf.NewKnowledgeBase(a)

	f := mustTmpFile("interfaces_")
	ifaces := netconf.NewIfacesConfig(d, f)
	log.Infof("reading template: %s", netconf.TplIfaces)
	tpl := mustRead(netconf.TplIfaces)
	mustApply(f, ifaces.Applier, tpl, "/etc/network/interfaces", false)
	_ = os.Remove(f)

	f = mustTmpFile("frr_")
	frr := netconf.NewFRRConfig(d, f)
	log.Infof("reading template: %s", netconf.TplFRR)
	tpl = mustRead(netconf.TplFRR)
	mustApply(f, frr.Applier, tpl, "/etc/frr/frr.conf", false)
	_ = os.Remove(f)

	f = mustTmpFile("rules.v4_")
	iptables := netconf.NewIptablesConfig(d, f)
	log.Infof("reading template: %s", netconf.TplIptables)
	tpl = mustRead(netconf.TplIptables)
	mustApply(f, iptables.Applier, tpl, "/etc/iptables/rules.v4", false)
	_ = os.Remove(f)

	chrony, err := system.NewChronyServiceEnabler(d)
	if err != nil {
		log.Warnf("failed to configure Chrony: %v", err)
	} else {
		err := chrony.Enable()
		if err != nil {
			log.Errorf("enabling Chrony failed: %v", err)
		}
	}

	log.Info("finished. Shutting down.")
}

func mustArg(index int) string {
	if len(os.Args) != 2 {
		log.Panic("expectation only the yaml input path is present as argument failed")
	}
	return os.Args[index]
}

func mustApply(tmpFile string, applier network.Applier, tpl string, dest string, reload bool) {
	t := template.Must(template.New(netconf.TplIfaces).Parse(tpl))
	err := applier.Apply(*t, tmpFile, dest, reload)
	if err != nil {
		log.Panic(err)
	}
}

func mustRead(name string) string {
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
