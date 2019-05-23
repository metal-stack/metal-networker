package main

import (
	"io/ioutil"
	"os"
	"text/template"

	"git.f-i-ts.de/cloud-native/metallib/zapup"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

const (
	// TplIfaces defines the name of the template to render interfaces configuration.
	TplIfaces = "interfaces.tpl"

	// TplFRR defines the name of the template to render FRR configuration.
	TplFRR = "frr.tpl"
)

var log = zapup.MustRootLogger().Sugar()

func main() {
	log.Info("running...")

	a := mustArg(1)
	log.Infof("loading: %s", a)
	d := NewKnowledgeBase(a)

	f := mustTmpFile("interfaces_")
	ifaces := NewIfacesConfig(d, f)
	log.Infof("reading template: %s", TplIfaces)
	tpl := mustRead(TplIfaces)
	log.Info("applying changes to /etc/network/interfaces")
	mustApply(f, ifaces.Applier, tpl, "/etc/network/interfaces")

	f = mustTmpFile("frr_")
	frr := NewFRRConfig(d, f)
	log.Infof("reading template: %s", TplFRR)
	tpl = mustRead(TplFRR)
	mustApply(f, frr.Applier, tpl, "/etc/network/interfaces")

	ioutil.

	log.Info("finished")
}

func mustArg(index int) string {
	if len(os.Args) != 2 {
		log.Panic("expectation only the yaml input path is present as argument failed")
	}
	return os.Args[index]
}

func mustApply(tmpFile string, applier network.Applier, tpl string, dest string) {
	t := template.Must(template.New(TplIfaces).Parse(tpl))
	err := applier.Apply(*t, tmpFile, dest)
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
	f, err := ioutil.TempFile("", prefix)
	if err != nil {
		log.Panic(err)
	}
	err = f.Close()
	if err != nil {
		log.Panic(err)
	}
	return f.Name()
}
