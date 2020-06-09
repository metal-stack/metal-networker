package netconf

import (
	// including template files with statik
	"io/ioutil"
	"text/template"

	_ "github.com/metal-stack/metal-networker/internal/netconf/tpl/statik"
	"github.com/rakyll/statik/fs"
)

func mustReadTpl(tplName string) string {
	statikFS, err := fs.New()
	if err != nil {
		log.Panic(err)
	}

	r, err := statikFS.Open("/" + tplName)
	if err != nil {
		log.Panic(err)
	}
	defer r.Close()

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		log.Panic(err)
	}

	return string(contents)
}

func mustParseTpl(tplName string) *template.Template {
	s := mustReadTpl(tplName)
	return template.Must(template.New(tplName).Parse(string(s)))
}
