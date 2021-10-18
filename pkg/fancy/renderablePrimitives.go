package fancy

import (
	"fmt"
	"html/template"
	"io"

	"github.com/ntaylor-barnett/gooey/pkg/register"
)

type TextRenderer struct {
	Value     string
	Class     string
	ClassText string
}

func NewTextPrimitve(val string) *TextRenderer {
	return &TextRenderer{
		Value: val,
	}
}

var Template_Text = `<span{{.ClassText}}>{{.Value}}</span>`

func (tr *TextRenderer) Write(ctx register.PageContext, w io.Writer) {
	t := template.Must(template.New("text").Parse(Template_Text))
	if tr.Class != "" {
		tr.ClassText = fmt.Sprintf(` class="%v"`, tr.Class)
	} else {
		tr.ClassText = ""
	}
	t.Execute(w, tr)
}
