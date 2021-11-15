package core

import (
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

// MakeRenderable will return a Renderable instance of whatever is given to it.
func MakeRenderablePrimitive(v interface{}) Renderable {
	switch vt := v.(type) {
	case Page:
		// a page being passed here is assumed to be a link (embedding is not allowed)
		l := NewLinkPrimitive(vt.Name(), "", vt)
		return l
	case register.PageStructure:
		// a page being passed here is assumed to be a link (embedding is not allowed)
		l := NewLinkPrimitive(vt.Title(), "", vt.Page())
		return l
	default:
		ref := reflect.ValueOf(v)
		for ref.Kind() == reflect.Ptr {
			ref = ref.Elem()
		}
		return NewTextPrimitve(fmt.Sprintf("%v", ref.Interface()))
	}

}

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
	// t := template.Must(template.New("link").Parse(Link_Text))
	u, err := ctx.ResolveUrl(lr.Destination)
	if err != nil {
		WriteComponentError(ctx, lr, err, w)
		return
	}
	// var tv struct {
	// 	DestURL template.URL
	// 	Attr    string
	// 	Value   string
	// }
	outString := Link_Text
	outString = strings.ReplaceAll(outString, "{{.URL}}", u.String())
	outString = strings.ReplaceAll(outString, "{{.Attr}}", createTagAttribs("class", lr.Class))
	outString = strings.ReplaceAll(outString, "{{.Value}}", lr.Text)
	_, err = w.Write([]byte(outString))
	if err != nil {
		WriteComponentError(ctx, lr, err, w)
		return
	}
}
