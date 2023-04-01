package core

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/ntaylor-barnett/gooey/pkg/htmlwriter"
	"github.com/ntaylor-barnett/gooey/register"
)

// the component container is a simple page that can display components. Each component supports the ability to render itself in the space.

// our basic Page interface. Ensures we meet the needs of the register system
type Page interface {
	register.Page
}

type ContainerPage struct {
	name       string
	components *LayoutComponent
}

func (cp *ContainerPage) WithName(n string) *ContainerPage {
	cp.name = n
	return cp
}

func (cp *ContainerPage) WithColumns(colCount int) *ContainerPage {
	if cp.components == nil {
		cp.components = NewLayoutComponent(colCount)
	} else {
		cp.components.columnCount = colCount
	}
	return cp

}

func (cp *ContainerPage) WithComponent(c Component) *ContainerPage {
	if cp.components == nil {
		cp.components = NewLayoutComponent(1)
	}
	cp.components.WithComponent(c)
	return cp
}

func (cp *ContainerPage) Name() string {
	return cp.name
}
func (cp *ContainerPage) Handler(ctx register.PageContext, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pw := newPageWriter(ctx, w)
		cp.components.WriteContent(ctx, pw)

	default:
		w.WriteHeader(405)
	}

}

func (cp *ContainerPage) OnHandlerAdded(parentPage register.Registerer) {
	// check to see if any of our components need to do some fancy registration
	cp.components.OnRegister(parentPage)
}

func WriteComponentError(ctx register.PageContext, c interface{}, err error, w io.Writer) {
	t, e := template.New("error").Parse(ErrTemplate)
	if e != nil {
		log.Fatal(e)
	}
	t.Execute(w, map[string]string{
		"Text": fmt.Sprintf("%v", err),
	})
}

func newPageWriter(ctx register.PageContext, w io.Writer) *pageWriter {
	return &pageWriter{
		Writer: w,
		ctx:    ctx,
		//regList: make(map[Component]RegisteredInfo),
	}
}

type pageWriter struct {
	io.Writer
	ctx register.PageContext
	//	regList map[Component]RegisteredInfo
}

func (pw *pageWriter) AddTextElement(string, ...func(htmlwriter.HtmlTextElement)) htmlwriter.HtmlNodeElement {
	return nil
}

// AddTemplatedHTML accepts either a "*html.Template" OR a string (to be interpreted as a template). The second parameter is the data to be parsed by the template
func (pw *pageWriter) AddTemplatedHTML(interface{}, interface{}) htmlwriter.HtmlNodeElement {
	return nil
}
func (pw *pageWriter) SetClosing(bool) htmlwriter.HtmlElement {
	return nil
}

func (pw *pageWriter) AddNodeElement(string, ...func(htmlwriter.HtmlNodeElement)) htmlwriter.HtmlNodeElement {
	return nil
}

// func (pw *pageWriter) RegisterComponent(c Component) RegisteredInfo {
// 	ix := len(pw.regList)
// 	compId := fmt.Sprintf("cid-%v", ix)
// 	ri := RegisteredInfo{
// 		Id: compId,
// 	}
// 	pw.regList[c] = ri
// 	return ri
// }
