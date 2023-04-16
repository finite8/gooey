package core

import (
	"io"

	"github.com/ntaylor-barnett/gooey/register"
)

// the LayoutComponent is simply a container that can contain other containers. It provides very simple layout capabilities.
type LayoutComponent struct {
	columnCount int
	children    []Component
}

// Parentable is a component that can have children
type ParentableComponent interface {
	GetChildren() []Component
}

func NewLayoutComponent(columnCount int) *LayoutComponent {
	return &LayoutComponent{
		columnCount: columnCount,
	}
}

func (cc *LayoutComponent) WithComponent(c Component) *LayoutComponent {
	cc.children = append(cc.children, c)
	return cc
}

func (cc *LayoutComponent) GetChildren() (ret []Component) {
	for _, c_iter := range cc.children {
		child := c_iter
		ret = append(ret, child)
		if p, ok := child.(ParentableComponent); ok {
			carr := p.GetChildren()
			if len(carr) > 0 {
				ret = append(ret, carr...)
			}
		}
	}
	return
}

func (cc *LayoutComponent) Write(ctx register.PageContext, w PageWriter) {
	io.WriteString(w, `<table>`)
	colPos := 0
	inRow := false
	for _, child := range cc.children {
		if inRow == false {
			io.WriteString(w, "<tr>")
			inRow = true
		}
		colPos++
		io.WriteString(w, `<td><div>`)
		child.Write(ctx, w)
		io.WriteString(w, `</div></td>`)
		if colPos == cc.columnCount {
			io.WriteString(w, "</tr>")
			inRow = false
			colPos = 0
		}
	}
	if inRow {
		io.WriteString(w, "</tr>")
	}

	io.WriteString(w, `</table>`)

}
func (cc *LayoutComponent) OnRegister(ctx register.Registerer) {
	for _, child := range cc.children {
		child.OnRegister(ctx)
	}
}
