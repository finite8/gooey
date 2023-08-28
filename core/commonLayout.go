package core

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/finite8/gooey/register"
)

const (
	leadingTemplate = `
<div id="GOOEY_header" class="GOOEY GOOEY_header">{{.Header}}</div>
<div id="GOOEY_status" class="GOOEY GOOEY_status">{{.Status}}</div>
<nav id="GOOEY_nav" class="GOOEY GOOEY_nav"><div class="GOOEY GOOEY_navcontent">{{.Nav}}</div></nav>
<div id="GOOEY_content" class="GOOEY GOOEY_content">`
	trailingTemplate = `</div>
<div id="GOOEY_footer" class="GOOEY GOOEY_footer">{{.Footer}}</div>
`
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

func (cl *CommonLayout) Render(ctx register.PageContext, w http.ResponseWriter, r *http.Request, pageRenderer func(ctx register.PageContext, w http.ResponseWriter, r *http.Request)) {

	cl.t.Lookup("leading").Execute(w, leadingData{
		Header: GetComponentHTML(ctx, cl.Header),
		Status: GetComponentHTML(ctx, cl.Status),
		Nav:    GetComponentHTML(ctx, cl.Nav),
	})
	pageRenderer(ctx, w, r)
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

func (cl *CommonLayout) QueryBehaviour(ctx register.PageContext, b register.Behaviour) register.Behaviour {
	return b
}
func GetComponentHTML(ctx register.PageContext, c Component) template.HTML {
	if c == nil {
		return template.HTML("")
	}
	w := &strings.Builder{}
	pw := newPageWriter(ctx, w)
	c.Write(ctx, pw)
	return template.HTML(w.String())
}
