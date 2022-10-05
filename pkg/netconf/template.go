package netconf

import (
	"embed"
	"path"
	"text/template"
)

//go:embed tpl
var templates embed.FS

func mustReadTpl(tplName string) string {
	contents, err := templates.ReadFile(path.Join("tpl", tplName))
	if err != nil {
		panic(err)
		return ""
	}
	return string(contents)
}

func mustParseTpl(tplName string) *template.Template {
	s := mustReadTpl(tplName)
	return template.Must(template.New(tplName).Parse(string(s)))
}
