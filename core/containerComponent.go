package core

import (
	"fmt"
	"io"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

// the LayoutComponent is simply a container that can contain other containers. It provides very simple layout capabilities.
type ContainerComponent struct {
	register.PageElement
	ComponentBase
	children []Component
	renderer ContainerRenderer
}

type ContainerRenderer interface {
	RenderContainer(cc *ContainerComponent, ctx register.PageContext, w PageWriter)
}

type basicRenderer struct {
}

func (br *basicRenderer) RenderContainer(cc *ContainerComponent, ctx register.PageContext, w PageWriter) {
	for _, child := range cc.children {
		io.WriteString(w, `<div>`)
		child.Write(ctx, w)
		io.WriteString(w, `</div>`)
	}
}

type CardRenderer struct {
	Header string
	Width  interface{}
}

func NewCardRenderer() *CardRenderer {
	return &CardRenderer{}
}

func (cr *CardRenderer) WithWidthPortion(p PortionSize) *CardRenderer {
	cr.Width = p
	return cr
}

func (cr *CardRenderer) WithHeader(txt string) *CardRenderer {
	cr.Header = txt
	return cr
}

type PortionSize string

const (
	Portion25  = PortionSize("25%")
	Portion50  = PortionSize("50%")
	Portion75  = PortionSize("75%")
	Portion100 = PortionSize("100%")
)

func (cr *CardRenderer) RenderContainer(cc *ContainerComponent, ctx register.PageContext, w PageWriter) {
	rootClasses := []string{"card"}
	if ps, ok := cr.Width.(PortionSize); ok {
		switch ps {
		case Portion25:
			rootClasses = append(rootClasses, "w-25")
		case Portion50:
			rootClasses = append(rootClasses, "w-50")
		case Portion75:
			rootClasses = append(rootClasses, "w-75")
		case Portion100:
			rootClasses = append(rootClasses, "w-100")
		}
	}
	io.WriteString(w, fmt.Sprintf(`<div class="%s">`, strings.Join(rootClasses, " ")))
	{
		io.WriteString(w, `<div class="card-body">`)
		if cr.Header != "" {
			io.WriteString(w, fmt.Sprintf(`<h4 class="card-title">%s</h4>`, cr.Header))
		}
		for _, child := range cc.children {
			io.WriteString(w, "<div>")
			child.Write(ctx, w)
			io.WriteString(w, `</div>`)
		}
		io.WriteString(w, `</div>`)

	}
	io.WriteString(w, `</div>`)
}

func NewContainerComponent() *ContainerComponent {
	return &ContainerComponent{
		PageElement: register.PageElement{
			Kind:        register.ElementTag_Closing,
			ElementName: "div",
		},
	}
}

func (cc *ContainerComponent) WithStyle(r ContainerRenderer) *ContainerComponent {
	cc.renderer = r
	return cc
}

func (cc *ContainerComponent) WithComponent(c Component) *ContainerComponent {
	cc.children = append(cc.children, c)
	return cc
}

func (cc *ContainerComponent) WithRenderables(r ...Renderable) {
	//TODO
}

func (cc *ContainerComponent) Write(ctx register.PageContext, w PageWriter) {

	if cc.renderer == nil {
		rndr := &basicRenderer{}
		rndr.RenderContainer(cc, ctx, w)
	} else {
		cc.renderer.RenderContainer(cc, ctx, w)
	}

}
func (cc *ContainerComponent) OnRegister(ctx register.Registerer) {
	for _, child := range cc.children {
		child.OnRegister(ctx)
	}
}
