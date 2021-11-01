package core

import (
	"html/template"
	"io"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

const (
	leadingTemplate = `<html>
<div id="GOOEY_header" class="GOOEY GOOEY_header">{{.Header}}</div>
<div id="GOOEY_status" class="GOOEY GOOEY_status">{{.Status}}</div>
<div id="GOOEY_nav" class="GOOEY GOOEY_nav">{{.Nav}}</div>
<div id="GOOEY_content" class="GOOEY GOOEY_content">`
	trailingTemplate = `</div>
<div id="GOOEY_footer" class="GOOEY GOOEY_footer">{{.Footer}}</div>
</html>`
)

type (
	leadingData struct {
		Header template.HTML
		Status template.HTML
		Nav    template.HTML
	}
	trailingData struct {
		Footer template.HTML
	}
)

// CommonLayout provides a flexible off-the-shelf layout for common site layouts. This gives you sections for:
// - Navigation Menus
// - Status
// - Headers
// - Footers
type CommonLayout struct {
	Nav    Component
	Status Component
	Header Component
	Footer Component
	t      *template.Template
}

func NewCommonLayout() *CommonLayout {
	cl := &CommonLayout{}
	t := template.Must(template.New("leading").Parse(leadingTemplate))
	t = template.Must(t.New("trailing").Parse(trailingTemplate))
	cl.t = t

	return cl

}

var _ register.PageLayout = (*CommonLayout)(nil)

func (cl *CommonLayout) RenderLeading(ctx register.PageContext, w io.Writer) {
	cl.t.Lookup("leading").Execute(w, leadingData{
		Header: GetComponentHTML(ctx, cl.Header),
		Status: GetComponentHTML(ctx, cl.Status),
		Nav:    GetComponentHTML(ctx, cl.Nav),
	})
}
func (cl *CommonLayout) RenderTrailing(ctx register.PageContext, w io.Writer) {
	cl.t.Lookup("trailing").Execute(w, trailingData{
		Footer: GetComponentHTML(ctx, cl.Footer),
	})
}
func (cl *CommonLayout) OnHandlerAdded(reg register.Registerer) {
	for _, c := range []Component{cl.Nav, cl.Status, cl.Header, cl.Footer} {
		if c != nil {
			c.OnRegister(reg)
		}
	}
}
func GetComponentHTML(ctx register.PageContext, c Component) template.HTML {
	if c == nil {
		return template.HTML("")
	}
	w := &strings.Builder{}
	c.WriteContent(ctx, w)
	return template.HTML(w.String())
}
