package htmlbuilder

import (
	"fmt"
	"io"
)

type htmlcontainer struct {
	tagname  string
	id       string
	children []HtmlNode
	attribs  *Attributable
}

func NewHtmlContainer(tag string) HtmlNode {
	return &htmlcontainer{
		tagname: tag,
	}
}

var _ HtmlNode = (*htmlcontainer)(nil)

func (c *htmlcontainer) AppendNode(newNode HtmlNode) {
	c.children = append(c.children, newNode)
}

func (c *htmlcontainer) Attributes() map[string]HtmlAttributeValue {
	if c.attribs == nil {
		c.attribs = NewAttributable()
	}
	return c.attribs.Attributes()
}

func (c *htmlcontainer) Name() string {
	return c.tagname
}

func (c *htmlcontainer) SetID(idVal string) {
	c.id = idVal
}

func (c *htmlcontainer) Write(w io.Writer) (n int, err error) {
	wrt := &appendingWriter{
		w: w,
	}
	defer func() {
		n = wrt.n
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()

	wrtArray := make([]Writable, len(c.children))
	for ix, v_iter := range c.children {
		v := v_iter
		wrtArray[ix] = v
	}
	WriteNodeElement(wrt, c.tagname, false, c.Attributes(), wrtArray)
	return
}
