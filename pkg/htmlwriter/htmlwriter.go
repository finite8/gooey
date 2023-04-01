package htmlwriter

import (
	"fmt"
	"html/template"
	"io"
	"strings"
)

type HtmlWriter struct {
	rootNode *node
	htmlCtx  *HtmlContext
}

func NewHtmlWriter() *HtmlWriter {
	return &HtmlWriter{
		htmlCtx: &HtmlContext{
			identifiedNodes: make(map[string]*node),
		},
	}
}

func (hw *HtmlWriter) WriteTo(w io.Writer) (int64, error) {
	return hw.rootNode.write(w)
}

func (hw *HtmlWriter) NewRoot(name string) HtmlNodeElement {
	if hw.rootNode != nil {
		panic("root already defined")
	}

	hw.rootNode = newNode(name, hw.htmlCtx)
	return hw.rootNode
}

type HtmlContext struct {
	identifiedNodes map[string]*node
}

func (hc *HtmlContext) AddId(id string, n *node) {
	_, ok := hc.identifiedNodes[id]
	if ok {
		// cannot replace an existing id
		panic(fmt.Sprintf("id %s is already in use", id))
	}
	hc.identifiedNodes[id] = n
}

type HtmlAttributeValue interface {
	GetValue() string                          // should be the text that is put into HTML
	ApplyValue(interface{}) HtmlAttributeValue // Apply the given value in a way that makes sense for the attribute.
	IsSet() bool                               // return true if it has a valid value and can be rendered. Return false otherwise
	MergeInto(HtmlAttributeValue)              // Copy the values from THIS attribute into the given one
}

type HtmlWritable interface {
	write(io.Writer) (int64, error)
}

type HtmlElement interface {
	HtmlWritable
	// SetAttribute will assign the attribute value into the given attribute name. If it already exists, this will merge the passed value into it.
	SetAttribute(string, HtmlAttributeValue) HtmlElement
	SetID(idVal string) HtmlElement
}

type HtmlTextElement interface {
	HtmlNodeElement
	AppendText(string) HtmlTextElement
}

type HtmlNodeElement interface {
	HtmlElement
	AddNodeElement(string, ...func(HtmlNodeElement)) HtmlNodeElement
	AddTextElement(string, ...func(HtmlTextElement)) HtmlNodeElement
	// AddTemplatedHTML accepts either a "*html.Template" OR a string (to be interpreted as a template). The second parameter is the data to be parsed by the template
	AddTemplatedHTML(interface{}, interface{}) HtmlNodeElement
	SetClosing(bool) HtmlElement
}

// the HtmlDocument represents a root HTML document, including all of its relevant elements.
type HtmlDocument interface {
	Head() HtmlElement
	Body() HtmlElement
}

type htmlDoc struct {
	doc  *HtmlWriter
	head HtmlElement
	body HtmlElement
}

func (hd *htmlDoc) Head() HtmlElement {
	if hd.head == nil {
		hd.head = hd.doc.rootNode.AddNodeElement("head", func(hne HtmlNodeElement) {})
	}
	return hd.head
}

type textWriter struct {
	text string
}

func (tw *textWriter) write(w io.Writer) (int64, error) {
	bn, e := w.Write([]byte(tw.text))
	return int64(bn), e
}

type node struct {
	htmlCtx   *HtmlContext
	name      string
	id        string
	children  []HtmlWritable
	attribs   map[string]HtmlAttributeValue
	isClosing bool
}

// a textnode has no child elements
type textnode struct {
	*node
	writableElements []HtmlWritable
}

func (tn *textnode) AppendText(s string) HtmlTextElement {
	tn.writableElements = append(tn.writableElements, &textWriter{
		text: s,
	})
	return tn
}

func (tn *textnode) AddNodeElement(name string, funcs ...func(HtmlNodeElement)) HtmlNodeElement {
	newNode := tn.newNodeElement(name, funcs...)
	tn.writableElements = append(tn.writableElements, newNode)
	return tn
}

func (tn *textnode) AddTextElement(name string, funcs ...func(HtmlTextElement)) HtmlNodeElement {
	newNode := tn.newTextElement(name, funcs...)
	tn.writableElements = append(tn.writableElements, newNode)
	return tn
}

func (tn *textnode) AddTemplatedHTML(t interface{}, p interface{}) HtmlNodeElement {
	newNode := tn.appendTemplatedHTML(t, p)
	tn.writableElements = append(tn.writableElements, newNode)
	return tn
}

func (tn *textnode) write(w io.Writer) (bn int64, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				retErr = v
			default:
				retErr = fmt.Errorf("%v", v)
			}
		}
	}()
	bn += mustWrite(w, fmt.Sprintf("<%s", tn.name)) // we always write this starting part of the element
	if tn.id != "" {
		bn += mustWrite(w, fmt.Sprintf(" ID=\"%s\"", tn.id))
	}

	bn += mustWrite(w, ">") // close the opening element tag
	for _, we := range tn.writableElements {
		if bytesWritten, err := we.write(w); err != nil {
			return 0, err
		} else {
			bn += bytesWritten
		}
	}

	bn += mustWrite(w, fmt.Sprintf("</%s>", tn.name))

	return bn, nil
}

func newNode(name string, hc *HtmlContext) *node {
	return &node{
		name:    name,
		attribs: make(map[string]HtmlAttributeValue),
		htmlCtx: hc,
	}
}
func (n *node) AddTemplatedHTML(t interface{}, p interface{}) HtmlNodeElement {
	_ = n.appendTemplatedHTML(t, p)
	return n
}

func (n *node) appendTemplatedHTML(tIn interface{}, p interface{}) *textWriter {
	sb := &strings.Builder{}
	var t *template.Template
	switch v := tIn.(type) {
	case *template.Template:
		t = v
	case string:
		t = template.Must(template.New("raw").Parse(v))
	case *string:
		t = template.Must(template.New("raw").Parse(*v))
	default:
		panic(fmt.Errorf("type %T is invalid for templating", tIn))
	}
	err := t.Execute(sb, p)
	if err != nil {
		panic(err)
	}
	rh := &textWriter{
		text: sb.String(),
	}
	n.children = append(n.children, rh)
	return rh

}
func (n *node) AddNodeElement(name string, funcs ...func(HtmlNodeElement)) HtmlNodeElement {
	_ = n.newNodeElement(name, funcs...)
	return n
}
func (n *node) newNodeElement(name string, funcs ...func(HtmlNodeElement)) *node {
	newNode := newNode(name, n.htmlCtx)
	n.children = append(n.children, newNode)
	for _, f := range funcs {
		if f != nil {
			f(newNode)
		}

	}
	return newNode

}
func (n *node) AddTextElement(name string, funcs ...func(HtmlTextElement)) HtmlNodeElement {
	_ = n.newTextElement(name, funcs...)
	return n
}
func (n *node) newTextElement(name string, funcs ...func(HtmlTextElement)) *textnode {
	newNode := newNode(name, n.htmlCtx)
	te := &textnode{
		node: newNode,
	}
	n.children = append(n.children, te)
	for _, f := range funcs {
		if f != nil {
			f(te)
		}

	}
	return te

}

func (n *node) SetAttribute(name string, a HtmlAttributeValue) HtmlElement {
	existing, ok := n.attribs[name]
	if ok {
		a.MergeInto(existing)
	} else {
		n.attribs[name] = a
	}
	return n
}

func (n *node) SetID(idVal string) HtmlElement {
	if n.id != "" {
		panic(fmt.Errorf("Id %s is already set on this node", n.id))
	}
	n.id = idVal
	return n
}

func (n *node) write(w io.Writer) (bn int64, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				retErr = v
			default:
				retErr = fmt.Errorf("%v", v)
			}
		}
	}()
	bn += mustWrite(w, fmt.Sprintf("<%s", n.name)) // we always write this starting part of the element
	if n.id != "" {
		bn += mustWrite(w, fmt.Sprintf(" ID=\"%s\"", n.id))
	}
	for name, attrib := range n.attribs {
		bn += mustWrite(w, fmt.Sprintf(" %s=\"%s\"", name, attrib.GetValue()))
	}
	if len(n.children) == 0 && n.isClosing {
		// a self-closing node that did not have any child elements
		bn += mustWrite(w, "/>") // close out the opening tag with an end slash and return as we are done
	} else {
		// otherwise, we have child elements OR we are not a self-closing node
		bn += mustWrite(w, ">") // close the opening element tag
		for _, child := range n.children {
			if bytesWrittem, e := child.write(w); e != nil {
				return bn, e
			} else {
				bn += bytesWrittem
			}

		}
		bn += mustWrite(w, fmt.Sprintf("</%s>", n.name))
	}
	return bn, nil
}

func mustWrite(w io.Writer, s string) int64 {
	n, e := w.Write([]byte(s))
	if e != nil {
		panic(e)
	}
	if n != len(s) {
		panic(fmt.Errorf("bytes written mismatch. Expected %v, Actual %v", len(s), n))
	}
	return int64(n)
}

func (n *node) SetClosing(c bool) HtmlElement {
	n.isClosing = c
	return n
}
