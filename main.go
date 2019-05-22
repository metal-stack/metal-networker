package main

import (
	"git.f-i-ts.de/cloud-native/metallib/network"
	"io/ioutil"
	"os"
	"text/template"
)

const (
	TplIfaces = "interfaces.tpl"
	TplFrr    = "frr.tpl"
)

func main() {
	a := os.Args[1]
	d, err := NewKnowledgeBase(a)
	if err != nil {
		panic(err)
	}

	f := mustTmpFile("interfaces_")
	ifaces := NewIfacesConfig(*d, f)
	tpl := mustRead(TplIfaces)
	mustApply(f, ifaces.Applier, tpl, "/etc/network/interfaces")

	f = mustTmpFile("frr_")
	frr := NewFrrConfig(*d, f)
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
	c, err := ioutil.ReadFile(TplIfaces)
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
