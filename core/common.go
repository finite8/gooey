package core

import (
	"fmt"
	"io"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

// the component is the most basic element that allows itself to be rendered.
type Component interface {
	WriteContent(register.PageContext, PageWriter)
	OnRegister(ctx register.Registerer)
	GetAttributes(ctx register.PageContext) string
}

type Renderable interface {
	Write(register.PageContext, io.Writer)
}

func WriteElements(ctx register.PageContext, prefix, suffix string, w PageWriter, values ...interface{}) {
	for _, v := range values {
		w.Write([]byte(prefix))
		WriteElement(ctx, w, v)
		w.Write([]byte(suffix))
	}
}

func WriteElement(ctx register.PageContext, w PageWriter, val interface{}) {
	if r, ok := val.(Renderable); ok {
		r.Write(ctx, w)
		return
	}
	if comp, ok := val.(Component); ok {
		comp.WriteContent(ctx, w)
		return
	}
	text := fmt.Sprintf("%v", val)
	NewTextPrimitve(text).Write(ctx, w)
}

func WriteComponent(ctx register.PageContext, w PageWriter, c Component) {

}

func createTagAttribs(vals ...string) string {
	var attrs []string
	for ix := 0; ix < len(vals); ix += 2 {
		attr := vals[ix]
		val := vals[ix+1]
		if strings.TrimSpace(val) != "" {
			attrs = append(attrs, fmt.Sprintf("%v=\"%v\"", attr, val))
		}
	}
	if len(attrs) > 0 {
		return fmt.Sprintf(" %v", strings.Join(attrs, " "))
	}
	return ""
}

type PageWriter interface {
	io.Writer
	//RegisterComponent(c Component) RegisteredInfo
}

type RegisteredInfo struct {
	Id string
}
