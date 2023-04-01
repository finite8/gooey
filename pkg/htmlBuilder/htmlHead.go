package htmlbuilder

import (
	"fmt"
	"io"
	"strings"
)

type Attributable struct {
	attribs map[string]HtmlAttributeValue
}

func (eb *Attributable) Attributes() map[string]HtmlAttributeValue {
	if eb.attribs == nil {
		eb.attribs = make(map[string]HtmlAttributeValue)
	}
	return eb.attribs
}

func (eb *Attributable) WithStringValues(strs ...string) *Attributable {
	l := len(strs)
	if l%2 == 1 {
		strs = append(strs, "")
	}
	for i := 0; i < len(strs); i += 2 {
		key := strs[i]
		value := strs[i+1]
		eb.Attributes()[key] = &StringAttribute{Value: value}
	}
	return eb
}

func NewAttributable() *Attributable {
	return &Attributable{}
}

type htmlhierachialelementbase struct {
	Attributable
	tagname      string
	endforbidden bool
	// this type of html element is one that supports having other elements. It does NOT take on the responsibility of ensuring uniqueness
	children []Writable

	innerBodyWriter func([]Writable, io.Writer) (int, error)
}

func newhtmlhierachialelementbase(name string, endforbid bool) *htmlhierachialelementbase {
	return &htmlhierachialelementbase{
		tagname:      name,
		endforbidden: endforbid,
	}
}

func WriteNodeElement(w io.Writer, tagname string, endforbidden bool, attribs map[string]HtmlAttributeValue, children []Writable) (bytesWritten int, err error) {
	// first, we need to write our tag
	tagToWrite := fmt.Sprintf("<%s", tagname)

	// so, we might have elements. Lets see if we have attributes next.
	attribsWritten := 0
	if len(attribs) > 0 {
		for k, v := range attribs {
			tagToWrite += fmt.Sprintf(" %s=\"%s\"", k, v.GetValue())
			attribsWritten++
		}
	}
	var (
		sbBytesWritten int
		n              int
	)

	// Next, we need to see if we have inner body to write.
	sb := &strings.Builder{}
	// we need to write the body ourselves.
	for _, child := range children {
		n, err = child.Write(sb)
		if err != nil {
			return
		}
		sbBytesWritten += n
	}
	if endforbidden {
		tagToWrite += ">"
	} else if sbBytesWritten > 0 {
		// some body text has been written, so we need to write out
		tagToWrite += ">" + sb.String() + fmt.Sprintf("</%s>", tagname)
	} else {
		// we can self close
		tagToWrite += "/>"
	}

	n, e := w.Write([]byte(tagToWrite))
	if e != nil {
		err = e
		return
	}
	bytesWritten += n
	return
}

func (eb *htmlhierachialelementbase) Write(w io.Writer) (int, error) {
	return WriteNodeElement(w, eb.tagname, eb.endforbidden, eb.attribs, eb.children)
}

func (eb *htmlhierachialelementbase) Name() string {
	return eb.tagname
}

type idablehtmlhierachialelementbase struct {
	*htmlhierachialelementbase
}

func (i *idablehtmlhierachialelementbase) SetID(idVal string) {
	i.htmlhierachialelementbase.Attributes()["id"] = &StringAttribute{Value: idVal}
}

type htmlhead struct {
	*htmlhierachialelementbase
	metaElements map[string]AttributableElement
}

func NewHtmlHead() *htmlhead {
	return &htmlhead{
		htmlhierachialelementbase: newhtmlhierachialelementbase("HEAD", false),
	}
}

func (h *htmlhead) Meta() map[string]AttributableElement {
	if h.metaElements == nil {
		h.metaElements = make(map[string]AttributableElement)
	}
	return h.metaElements
}

var _ HtmlHead = (*htmlhead)(nil)

func (h *htmlhead) Write(w io.Writer) (int, error) {
	var metaelements []Writable
	for k, v := range h.Meta() {
		newMetaItem := newhtmlhierachialelementbase("meta", true)
		v.Attributes()["name"] = &StringAttribute{Value: k}
		for attribName, attribValue := range v.Attributes() {
			newMetaItem.Attributes()[attribName] = attribValue
		}
		metaelements = append(metaelements, newMetaItem)
	}
	return WriteNodeElement(w, h.tagname, h.endforbidden, h.attribs, append(metaelements, h.children...))
}
