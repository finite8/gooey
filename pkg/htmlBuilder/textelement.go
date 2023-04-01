package htmlbuilder

import "io"

// only writes raw text out. Nothing else fancy happens
type textwritable struct {
	innertext string
}

func (tw *textwritable) Write(w io.Writer) (int, error) {
	return w.Write([]byte(tw.innertext))
}

// Basic textual element. i.e: <p attrib="soemthing">text</p>
type textelement struct {
	tagname   string
	innertext string
	attribs   map[string]HtmlAttributeValue
}

func (te *textelement) Name() string {
	return te.tagname
}

func (te *textelement) Attributes() map[string]HtmlAttributeValue {
	if te.attribs == nil {
		te.attribs = map[string]HtmlAttributeValue{}
	}
	return te.attribs
}
func (te *textelement) GetString() string {
	return te.innertext
}
func (te *textelement) SetString(s string) {
	te.innertext = s
}
func (te *textelement) Write(w io.Writer) (int, error) {
	return WriteNodeElement(w, te.tagname, false, te.attribs, []Writable{&textwritable{innertext: te.innertext}})
}
