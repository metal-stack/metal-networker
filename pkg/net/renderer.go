package net

import (
	"html/template"
	"io"
)

// Renderer is a thing to render content.
type Renderer struct {
}

// Render renders the given template using the given data into the provided writer.
func (r *Renderer) Render(w io.Writer, data interface{}, tpl template.Template) error {
	return tpl.Execute(w, data)
}
