package htmlbuilder

import (
	"fmt"
	"io"
)

type Writable interface {
	Write(io.Writer) (int, error)
}

type HtmlDoc interface {
	Writable
	Version() StringItem
	// no HTML element needed as it is an automatic component of any document
	// a HTML document cannot be a fragment as a HTML document is for describing the whole thing. a Fragment however can be one of the added nodes.
	Head() HtmlHead
	Title() TextElement
	Body() HtmlNode
}

type StringItem interface {
	GetString() string
	SetString(string)
}

// ==================================================
// The *Element interfaces define the core behaviours that other things can implement

type TagElement interface {
	// Name cannot be changed once declared
	Name() string
}

// TextElement is a basic node that does not support child elements
type TextElement interface {
	Writable
	TagElement
	AttributableElement
	StringItem
}

// This type of element supports specifying attributes an attribute cannot appear more than once, so they are presented as maps.
type AttributableElement interface {
	Attributes() map[string]HtmlAttributeValue
}

type IdentifiableElement interface {
	SetID(idVal string)
}

type HtmlNodeContainer interface {
	AppendNode(newNode HtmlNode)
}

// ==================================================
// The Html* Prefix defines what we actually put into our documents

// A NodeElement supports a hierachy of other nodes in its structure
type HtmlNode interface {
	Writable
	TagElement
	AttributableElement
	IdentifiableElement
	HtmlNodeContainer
}

type HtmlHead interface {
	Writable
	TagElement
	AttributableElement
	// Meta is an array of "meta" elements. These do not have closing tags (end tag forbidden)
	Meta() map[string]AttributableElement
}

type htmldoc struct {
	head  HtmlHead
	title TextElement
	ver   StringItem
	body  HtmlNode
}

func (hd *htmldoc) Head() HtmlHead {
	if hd.head == nil {
		hd.head = NewHtmlHead()
	}
	return hd.head
}

func (hd *htmldoc) Title() TextElement {
	if hd.title == nil {
		hd.title = &textelement{tagname: "TITLE"}
	}
	return hd.title
}

func (hd *htmldoc) Version() StringItem {
	if hd.ver == nil {
		hd.ver = &textelement{}
	}
	return hd.ver
}

func (hd *htmldoc) Body() HtmlNode {
	if hd.body == nil {
		hd.body = NewHtmlContainer("BODY")
	}
	return hd.body
}

func (hd *htmldoc) Write(w io.Writer) (n int, err error) {
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
	// lets write this document
	wrt.MustWrite("<HTML>")
	{
		wrt.MustWrite(hd.Head())
		wrt.MustWrite(hd.Title())
		if hd.body != nil {
			wrt.MustWrite(hd.body)
		}
	}
	wrt.MustWrite("</HTML>")
	return
}

var _ HtmlDoc = (*htmldoc)(nil)

type appendingWriter struct {
	n int
	w io.Writer
}

// // Writes the given string to the writer. Will panic if it encounters an error
// func (aw *appendingWriter) MustWriteString(s string) {
// 	_, err := aw.Write([]byte(s))
// 	if err != nil {
// 		panic(err)
// 	}
// }

// Writes the given writable to the writer. Will panic if it encounters an error
func (aw *appendingWriter) MustWrite(initem interface{}) {
	switch item := initem.(type) {
	case Writable:
		_, e := item.Write(aw) // we don't need to update n as the "Write" method does that for us
		if e != nil {
			panic(e)
		}
	case string:
		_, err := aw.Write([]byte(item))
		if err != nil {
			panic(err)
		}
	case []byte:
		_, err := aw.Write(item)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Sprintf("don't know how to write %T", initem))
	}

}

func (aw *appendingWriter) Write(p []byte) (int, error) {
	n, e := aw.w.Write(p)
	aw.n += n
	return n, e
}
