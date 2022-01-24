package core

import (
	"io"

	"github.com/ntaylor-barnett/gooey/register"
)

// the LayoutComponent is simply a container that can contain other containers. It provides very simple layout capabilities.
type ContainerComponent struct {
	register.PageElement
	children []Component
}

func NewContainerComponent() *ContainerComponent {
	return &ContainerComponent{
		PageElement: register.PageElement{
			Kind:        register.ElementTag_Closing,
			ElementName: "div",
		},
	}
}

func (cc *ContainerComponent) WithComponent(c Component) *ContainerComponent {
	cc.children = append(cc.children, c)
	return cc
}

func (cc *ContainerComponent) WriteContent(ctx register.PageContext, w PageWriter) {

	// if cc.GetKind() != register.ElementTag_Closing {
	// 	WriteComponentError(ctx, cc, ,w)
	// }

	for _, child := range cc.children {

		io.WriteString(w, `<div>`)
		child.WriteContent(ctx, w)
		io.WriteString(w, `</div>`)

	}

}
func (cc *ContainerComponent) OnRegister(ctx register.Registerer) {
	for _, child := range cc.children {
		child.OnRegister(ctx)
	}
}
