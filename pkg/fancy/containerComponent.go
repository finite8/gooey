package fancy

import (
	"io"

	"example.com/goboganui/pkg/register"
)

// the ContainerComponent is simply a container that can contain other containers. It provides very simple layout capabilities.
type ContainerComponent struct {
	columnCount int
	children    []Component
}

func NewContainerComponent(columnCount int) *ContainerComponent {
	return &ContainerComponent{
		columnCount: columnCount,
	}
}

func (cc *ContainerComponent) WithComponent(c Component) *ContainerComponent {
	cc.children = append(cc.children, c)
	return cc
}

func (cc *ContainerComponent) WriteContent(ctx register.PageContext, w io.Writer) {
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
		child.WriteContent(ctx, w)
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
func (cc *ContainerComponent) OnRegister(ctx register.Registerer) {
	for _, child := range cc.children {
		child.OnRegister(ctx)
	}
}
