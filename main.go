package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/metal-pod/v"

	"git.f-i-ts.de/cloud-native/metallib/zapup"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

const (
	// TplIfaces defines the name of the template to render interfaces configuration.
	TplIfaces = "interfaces.tpl"

	// TplFRR defines the name of the template to render FRR configuration.
	TplFRR = "frr.tpl"

	// TplIptables defines the name of the template to render iptables configuration.
	TplIptables = "rules.v4.tpl"
)

var log = zapup.MustRootLogger().Sugar()

func main() {
	log.Infof("running app version: %s", v.V.String())

	a := mustArg(1)
	log.Infof("loading: %s", a)
	d := NewKnowledgeBase(a)

	f := mustTmpFile("interfaces_")
	ifaces := NewIfacesConfig(d, f)
	log.Infof("reading template: %s", TplIfaces)
	tpl := mustRead(TplIfaces)
	mustApply(f, ifaces.Applier, tpl, "/etc/network/interfaces", false)
	_ = os.Remove(f)

	f = mustTmpFile("frr_")
	frr := NewFRRConfig(d, f)
	log.Infof("reading template: %s", TplFRR)
	tpl = mustRead(TplFRR)
	mustApply(f, frr.Applier, tpl, "/etc/frr/frr.conf", false)
	_ = os.Remove(f)

	f = mustTmpFile("rules.v4_")
	iptables := NewIptablesConfig(d, f)
	log.Infof("reading template: %s", TplIptables)
	tpl = mustRead(TplIptables)
	mustApply(f, iptables.Applier, tpl, "/etc/iptables/rules.v4", false)
	_ = os.Remove(f)

	log.Info("finished. Shutting down.")
}

func mustArg(index int) string {
	if len(os.Args) != 2 {
		log.Panic("expectation only the yaml input path is present as argument failed")
	}
	return os.Args[index]
}

func mustApply(tmpFile string, applier network.Applier, tpl string, dest string, reload bool) {
	t := template.Must(template.New(TplIfaces).Parse(tpl))
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

func versionHeader(uuid string) string {
	return fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n# Do not edit.",
		uuid, v.V.String())
}
