package core

import (
	"html/template"
	"io"

	"github.com/ntaylor-barnett/gooey/register"
)

type TextRenderer struct {
	Value     string
	Class     string
	classText string
}

func NewTextPrimitve(val string) *TextRenderer {
	return &TextRenderer{
		Value: val,
	}
}

var Template_Text = `<span{{.Attr}}>{{.Value}}</span>`

func (tr *TextRenderer) Write(ctx register.PageContext, w io.Writer) {
	t := template.Must(template.New("text").Parse(Template_Text))

	t.Execute(w, map[string]string{
		"Attr":  createTagAttribs("class", tr.Class),
		"Value": tr.Value,
	})
}

var Link_Text = `<a href="{{.URL}}"{{.Attr}}>{{.Value}}</a>`

type LinkRenderer struct {
	Text        string
	Destination interface{}
	Target      string
	Class       string
	classText   string
}

func NewLinkPrimitive(text, target string, dest interface{}) *LinkRenderer {
	return &LinkRenderer{
		Text:        text,
		Target:      target,
		Destination: dest,
	}
}

func (lr *LinkRenderer) Write(ctx register.PageContext, w io.Writer) {
	t := template.Must(template.New("text").Parse(Template_Text))
	u, err := ctx.ResolveUrl(lr.Destination)
	if err != nil {
		WriteComponentError(ctx, lr, err, w)
		return
	}
	var tv struct {
		DestURL template.URL
		Attr    string
		Value   string
	}
	tv.Attr = createTagAttribs("class", lr.Class)
	tv.DestURL = template.URL(u.String())
	tv.Value = lr.Text
	t.Execute(w, tv)
}
