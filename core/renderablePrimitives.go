package core

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

type nilComponent struct {
}

func (nc *nilComponent) OnRegister(ctx register.Registerer) {
	// this has nothing to do on register
}

func (nc *nilComponent) setRenderState(s renderstate) {

}

// MakeRenderable will return a Renderable instance of whatever is given to it.
func MakeRenderablePrimitive(v interface{}) Renderable {
	return makeRenderableInternal(v, RenderOption{}, 0)
}

type RenderOption struct {
	// Maximum depth of child elements this should render (default 6)
	MaxDepth *uint
	// What depth should a collapse should start being rendered (default 0)
	CollapseStart *uint
	// What depth should a collapse stop being rendered. If 0, all depths after start is rendered (default 0)
	CollapseEnd *uint
}

var (
	default_MaxDept       = uint(6)
	default_CollapseStart = uint(0)
	default_CollapseEnd   = uint(0)
)

func applyDefaults(v RenderOption) RenderOption {
	ret := RenderOption{
		MaxDepth:      &default_MaxDept,
		CollapseStart: &default_CollapseStart,
		CollapseEnd:   &default_CollapseEnd,
	}
	if v.MaxDepth != nil {
		ret.MaxDepth = v.MaxDepth
	}
	if v.CollapseStart != nil {
		ret.CollapseStart = v.CollapseStart
	}
	if v.CollapseEnd != nil {
		ret.CollapseEnd = v.CollapseEnd
	}
	return ret
}

func compileRenderOptions(opts []RenderOption) (ret RenderOption) {
	ret = RenderOption{
		MaxDepth:      &default_MaxDept,
		CollapseStart: &default_CollapseStart,
		CollapseEnd:   &default_CollapseEnd,
	}
	for _, v := range opts {
		if v.MaxDepth != nil {
			ret.MaxDepth = v.MaxDepth
		}
		if v.CollapseStart != nil {
			ret.CollapseStart = v.CollapseStart
		}
		if v.CollapseEnd != nil {
			ret.CollapseEnd = v.CollapseEnd
		}
	}
	return
}

// MakeRenderable makes a renderable view of the given interface. If "collapse" is true, anything that would
// render as something large will
func MakeRenderable(v interface{}, opts ...RenderOption) Renderable {
	// struct reflection is predictable and deterministic, maps are random so will sort by key
	opt := compileRenderOptions(opts)
	return makeRenderableInternal(v, opt, 0)
}

type renderstate struct {
	opt  RenderOption
	dept int
}

func makeRenderableInternal(v interface{}, opt RenderOption, currDepth int) Renderable {
	opt = applyDefaults(opt)
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
		var val interface{}
		if v == nil {
			val = nil
		} else {
			ref := reflect.ValueOf(v)
			for ref.Kind() == reflect.Ptr {
				ref = ref.Elem()
			}

			val = ref.Interface()
			switch ref.Kind() {
			case reflect.Struct, reflect.Map:
				if currDepth > int(*opt.MaxDepth) {
					return NewTextPrimitve(fmt.Sprintf("%v", val))
				}
				// if it is a map or an object, our render approach needs to be a bit more powerful
				oc := NewObjectComponent(func(pc register.PageContext) (interface{}, error) {
					return val, nil
				})
				oc.setRenderState(renderstate{
					opt:  opt,
					dept: currDepth + 1,
				})
				if int(*opt.CollapseStart) <= currDepth && (*opt.CollapseEnd == 0 || int(*opt.CollapseEnd) > currDepth) {
					return MakeExpandable(oc)
				} else {
					return oc
				}

			case reflect.Array, reflect.Slice:
				if currDepth > int(*opt.MaxDepth) {
					return NewTextPrimitve(fmt.Sprintf("%v", val))
				}
				tc := NewTableComponent(func(pc register.PageContext) (interface{}, error) {
					return val, nil
				})
				tc.setRenderState(renderstate{
					opt:  opt,
					dept: currDepth + 1,
				})
				if int(*opt.CollapseStart) <= currDepth && (*opt.CollapseEnd == 0 || int(*opt.CollapseEnd) > currDepth) {
					return MakeExpandable(tc)
				} else {
					return tc
				}

			}
		}
		return NewTextPrimitve(fmt.Sprintf("%v", val))
	}
}

// func makeMapRenderable(m map[string]interface{}) Renderable {

// }

func MakeExpandable(r Renderable) Renderable {
	rw := &RenderWrapper{
		f: func(pc register.PageContext, pw PageWriter) {
			seq := pc.GetNewSequence()
			collapseId := fmt.Sprintf("expandable%d", seq)
			buttonArea := NewTag("p", nil, func() Renderable {
				return NewTag("a", map[string]interface{}{
					"class":          "btn btn-primary",
					"data-bs-toggle": "collapse",
					"href":           fmt.Sprintf("#%s", collapseId),
					"role":           "button",
					"aria-expanded":  false,
					"aria-controls":  collapseId,
				}, "expand")
			})
			collapseArea := NewTag("div", map[string]interface{}{
				"class": "collapse",
				"id":    collapseId,
			}, NewTag("div", map[string]interface{}{
				"class": "card card-body",
			}, r))
			buttonArea.Write(pc, pw)
			collapseArea.Write(pc, pw)

		},
	}
	return rw
}

type RenderWrapper struct {
	f func(register.PageContext, PageWriter)
}

func (rw *RenderWrapper) Write(ctx register.PageContext, pw PageWriter) {
	rw.f(ctx, pw)
}

type TextRenderer struct {
	nilComponent
	Value     string
	Class     string
	classText string
}

func NewTextPrimitve(val string) *TextRenderer {
	return &TextRenderer{
		Value: val,
	}
}

var Template_Text = `<span {{.Attr}}>{{.Value}}</span>`

func (tr *TextRenderer) Write(ctx register.PageContext, w PageWriter) {
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

func (lr *LinkRenderer) Write(ctx register.PageContext, w PageWriter) {
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
