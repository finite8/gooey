package fancy

import (
	"fmt"
	"io"

	"github.com/ntaylor-barnett/gooey/pkg/register"
)

// the component is the most basic element that allows itself to be rendered.
type Component interface {
	WriteContent(register.PageContext, io.Writer)
	OnRegister(ctx register.Registerer)
}

type Renderable interface {
	Write(register.PageContext, io.Writer)
}

func WriteElements(ctx register.PageContext, prefix, suffix string, w io.Writer, values ...interface{}) {
	for _, v := range values {
		w.Write([]byte(prefix))
		WriteElement(ctx, w, v)
		w.Write([]byte(suffix))
	}
}

func WriteElement(ctx register.PageContext, w io.Writer, val interface{}) {
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
