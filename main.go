package main

import (
	"io/ioutil"
	"os"
	"text/template"
)

const (
	TplIfaces = "interfaces.tpl"
)

func main() {
	// we need a preprocessor that aggregates (e.g. skip tenant child networks)/ prepares (split prefix/ length)
	a := os.Args[1]
	d, err := NewKnowledgeBase(a)
	if err != nil {
		panic(err)
	}

	f := mustTmpFile("interfaces_")
	c := NewIfacesConfig(*d, f)
	tpl := mustRead(TplIfaces)
	t := template.Must(template.New(TplIfaces).Parse(tpl))
	c.Applier.Apply(*t, f, "/etc/network/interfaces")
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
