package htmlwriter

import "io"

type HtmlWriter struct {
	w         io.Writer
	isRoot    bool
	isClosing bool
	attribs   map[string]string
}

type HtmlElement interface {
	WriteElement(string, func(HtmlElement) error)
	SetAttribute(attrName string, attrValue string) HtmlElement
}

func NewHtmlWriter(w io.Writer) HtmlElement {
	return &HtmlWriter{
		w:         w,
		isRoot:    true,
		isClosing: false,
	}
}

func (h *HtmlWriter) WriteElement(name string, f func(HtmlElement) error) {
	// so this will not start writing until its child element is complete
}
