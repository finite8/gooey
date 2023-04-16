package core

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

// the component container is a simple page that can display components. Each component supports the ability to render itself in the space.

// our basic Page interface. Ensures we meet the needs of the register system
type Page interface {
	register.Page
}

type ContainerPage struct {
	name               string
	components         *LayoutComponent
	postableComponents []PostableComponent
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
func (cp *ContainerPage) Handler(ctx register.PageContext, w http.ResponseWriter, r *http.Request) interface{} {
	var pw *pageWriter
	switch r.Method {
	case http.MethodGet:
		pw = newPageWriter(ctx, w)
		cp.components.Write(ctx, pw)
	case http.MethodPost:
		// we now need to go through all of our post handlers to see if something needs to be done.
		isHandled := false
		for _, h := range cp.postableComponents {
			pr := h.HandlePost(ctx, r)
			if pr.IsHandled {
				isHandled = true
			}
			if pr.HaltProcessing {
				break
			}
		}
		if isHandled {
			// the post has been handled by a component. We can continue rendering
			pw = newPageWriter(ctx, w)
			cp.components.Write(ctx, pw)
		} else {
			WriteComponentError(ctx, nil, errors.New("the POST data wa either invalid or not handled by any component"), w)
			w.WriteHeader(400)
			return nil
		}

	default:
		w.WriteHeader(405)
	}

	return nil

}

func (cp *ContainerPage) OnHandlerAdded(parentPage register.Registerer) {
	// check to see if any of our components need to do some fancy registration
	cp.components.OnRegister(parentPage)
	var postable []PostableComponent
	for _, comp := range cp.components.GetChildren() {
		switch item := comp.(type) {
		case PostableComponent:
			postable = append(postable, item)
		}
	}
	cp.postableComponents = postable
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

type scriptBlock struct {
	name       string
	scriptType string
	data       strings.Builder
}

func (sb *scriptBlock) Write(p []byte) (n int, err error) {
	return sb.data.Write(p)
}

func (sb *scriptBlock) WriteString(s string) (n int, err error) {
	return sb.data.WriteString(s)
}

type pageWriter struct {
	io.Writer
	ctx     register.PageContext
	scripts []*scriptBlock
	//	regList map[Component]RegisteredInfo
}

func (pw *pageWriter) Finalize() {
	for _, sb := range pw.scripts {

		pw.Write([]byte(fmt.Sprintf(`<script type="%s">`, sb.scriptType) + "\n"))
		pw.Write([]byte(sb.data.String()))
		pw.Write([]byte("\n</script>\n"))
	}
}

func (pw *pageWriter) GetScriptWriter(scriptSectionName, scriptType string) ScriptWriter {
	for _, v := range pw.scripts {
		if v.name == scriptSectionName && v.scriptType == scriptType {
			return v
		}
	}
	sw := &scriptBlock{
		name:       scriptSectionName,
		scriptType: scriptType,
	}
	pw.scripts = append(pw.scripts, sw)
	return sw
}

func (pw *pageWriter) WriteElement(ctx register.PageContext, val interface{}) {
	WriteElement(ctx, pw, val)
}

// func (pw *pageWriter) AddTextElement(string, ...func(htmlwriter.HtmlTextElement)) htmlwriter.HtmlNodeElement {
// 	return nil
// }

// // AddTemplatedHTML accepts either a "*html.Template" OR a string (to be interpreted as a template). The second parameter is the data to be parsed by the template
// func (pw *pageWriter) AddTemplatedHTML(interface{}, interface{}) htmlwriter.HtmlNodeElement {
// 	return nil
// }
// func (pw *pageWriter) SetClosing(bool) htmlwriter.HtmlElement {
// 	return nil
// }

// func (pw *pageWriter) AddNodeElement(string, ...func(htmlwriter.HtmlNodeElement)) htmlwriter.HtmlNodeElement {
// 	return nil
// }

// func (pw *pageWriter) RegisterComponent(c Component) RegisteredInfo {
// 	ix := len(pw.regList)
// 	compId := fmt.Sprintf("cid-%v", ix)
// 	ri := RegisteredInfo{
// 		Id: compId,
// 	}
// 	pw.regList[c] = ri
// 	return ri
// }
