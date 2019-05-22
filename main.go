package main

import (
	"io/ioutil"
	"os"
	"text/template"

	"git.f-i-ts.de/cloud-native/metallib/network"
)

const (
	TplIfaces = "interfaces.tpl"
	TplFrr    = "frr.tpl"
)

func main() {
	// Todo: rethink: panic
	// TODO: ensure the arg is present
	a := os.Args[1]
	d := NewKnowledgeBase(a)

	f := mustTmpFile("interfaces_")
	ifaces := NewIfacesConfig(d, f)
	tpl := mustRead(TplIfaces)
	mustApply(f, ifaces.Applier, tpl, "/etc/network/interfaces")

	f = mustTmpFile("frr_")
	frr := NewFrrConfig(d, f)
	tpl = mustRead(TplFrr)
	mustApply(f, frr.Applier, tpl, "/etc/network/interfaces")
}

func mustApply(tmpFile string, applier network.Applier, tpl string, dest string) {
	t := template.Must(template.New(TplIfaces).Parse(tpl))
	err := applier.Apply(*t, tmpFile, dest)
	if err != nil {
		panic(err)
	}
}

func mustRead(name string) string {
	c, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(c)
}

func mustTmpFile(prefix string) string {
	f, err := ioutil.TempFile("", prefix)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
	return f.Name()
}
