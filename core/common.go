package core

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

// the component is the most basic element that allows itself to be rendered.
type Component interface {
	WriteContent(register.PageContext, PageWriter)
	OnRegister(ctx register.Registerer)
	GetAttributes(ctx register.PageContext) string
}

// A component that implements "HandlePost" will also be asked if it has accepted a POST event. If it has, then the normal page rendering will continue
type PostableComponent interface {
	HandlePost(register.PageContext, *http.Request) PostHandlerResult
}

type PostHandlerResult struct {
	// set to true if the post event has been handled. At least one component must handle the post for it to be accepted
	IsHandled bool
	// set to true if calling HandlePost should stop.
	HaltProcessing bool
}

type Renderable interface {
	Write(register.PageContext, PageWriter)
}

func WriteElements(ctx register.PageContext, prefix, suffix string, w PageWriter, values ...interface{}) {
	for _, v := range values {
		w.Write([]byte(prefix))
		WriteElement(ctx, w, v)
		w.Write([]byte(suffix))
	}
}

func WriteElement(ctx register.PageContext, w PageWriter, val interface{}) {
	switch v := val.(type) {
	case Renderable:
		v.Write(ctx, w)
	case []Renderable:
		for _, item := range v {
			item.Write(ctx, w)
		}
	case Component:
		v.WriteContent(ctx, w)
	default:
		text := fmt.Sprintf("%v", v)
		NewTextPrimitve(text).Write(ctx, w)
	}
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
	// GetScriptWriter returns a writer to allow you to write into a custom script block.
	GetScriptWriter(scriptSectionName, typename string) ScriptWriter
	//RegisterComponent(c Component) RegisteredInfo
	WriteElement(register.PageContext, interface{})
}

type ScriptWriter interface {
	io.Writer
	io.StringWriter
}

type RegisteredInfo struct {
	Id string
}

type TagPrimitive struct {
	Name       string
	Unpaired   bool
	Attributes map[string]interface{}
	InnerText  interface{}
}

func (tp *TagPrimitive) Write(ctx register.PageContext, w PageWriter) {
	w.Write([]byte(fmt.Sprintf("<%s", tp.Name)))
	if tp.Attributes != nil && len(tp.Attributes) > 0 {
		for k, v := range tp.Attributes {
			w.Write([]byte(" " + k))
			if v != nil {

				var textToWrite string
				switch aVal := v.(type) {
				case string:
					textToWrite = fmt.Sprintf(`"%s"`, aVal)
				case *string:
					textToWrite = fmt.Sprintf(`"%s"`, *aVal)
				default:
					rv := reflect.ValueOf(aVal)
					for rv.Kind() == reflect.Pointer {
						rv = rv.Elem()
					}
					textToWrite = fmt.Sprintf("%v", rv.Interface())
				}
				w.Write([]byte(fmt.Sprintf("=%s", textToWrite)))
			}
		}
	}
	if tp.Unpaired {
		// there cannot be inner text for this
		w.Write([]byte(">"))
	} else {
		if tp.InnerText != nil {
			WriteElement(ctx, w, tp.InnerText)
			w.Write([]byte(fmt.Sprintf("</%s>", tp.Name)))
		} else {
			w.Write([]byte("/>"))
		}
	}
}

func NewUnpairedTag(tagname string, attrib map[string]interface{}) *TagPrimitive {
	return &TagPrimitive{
		Name:       tagname,
		Unpaired:   true,
		Attributes: attrib,
	}
}

func NewTag(tagname string, attrib map[string]interface{}, inner interface{}) *TagPrimitive {
	tp := &TagPrimitive{
		Name:       tagname,
		Unpaired:   false,
		Attributes: attrib,
	}
	switch i := inner.(type) {
	case func() Renderable:
		tp.InnerText = i()
	case func() []Renderable:
		tp.InnerText = i()
	default:
		tp.InnerText = i
	}
	return tp
}
