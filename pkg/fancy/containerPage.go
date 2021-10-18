package fancy

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"example.com/goboganui/pkg/register"
)

// the component container is a simple page that can display components. Each component supports the ability to render itself in the space.

// our basic Page interface. Ensures we meet the needs of the register system
type Page interface {
	register.Page
}

type ContainerPage struct {
	name       string
	components *ContainerComponent
}

func (cp *ContainerPage) WithName(n string) *ContainerPage {
	cp.name = n
	return cp
}

func (cp *ContainerPage) WithColumns(colCount int) *ContainerPage {
	if cp.components == nil {
		cp.components = NewContainerComponent(colCount)
	} else {
		cp.components.columnCount = colCount
	}
	return cp

}

func (cp *ContainerPage) WithComponent(c Component) *ContainerPage {
	if cp.components == nil {
		cp.components = NewContainerComponent(1)
	}
	cp.components.WithComponent(c)
	return cp
}

func (cp *ContainerPage) Name() string {
	return cp.name
}
func (cp *ContainerPage) Handler(ctx register.PageContext, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<html>"))
	cp.components.WriteContent(ctx, w)
	w.Write([]byte("</html>"))
}

func (cp *ContainerPage) OnHandlerAdded(parentPage register.Registerer) {
	// check to see if any of our components need to do some fancy registration
	cp.components.OnRegister(parentPage)
}

func WriteComponentError(ctx register.PageContext, c Component, err error, w io.Writer) {
	t, e := template.New("error").Parse(ErrTemplate)
	if e != nil {
		log.Fatal(e)
	}
	t.Execute(w, map[string]string{
		"Text": fmt.Sprintf("%v", err),
	})
}