package htmlwriter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeWriting(t *testing.T) {
	hw := NewHtmlWriter()
	hw.NewRoot("html").
		AddNodeElement("head").
		AddTextElement("body", func(hte HtmlTextElement) {
			hte.AppendText("text outside").AddTextElement("p", func(hte HtmlTextElement) { hte.AppendText("text inside") })
		}).
		AddNodeElement("footer", func(hne HtmlNodeElement) {
			hne.AddTextElement("one")
			hne.AddNodeElement("two", func(hne HtmlNodeElement) { hne.SetClosing(true) })
		})
	sb := &strings.Builder{}
	n, err := hw.WriteTo(sb)
	assert.Nil(t, err)
	assert.Equal(t, int64(sb.Len()), n)
	assert.Equal(t, "<html><head></head><body>text outside<p>text inside</p></body><footer><one></one><two/></footer></html>", sb.String())
}

func TestTextNodeFunctions(t *testing.T) {
	pl := struct {
		Text string
	}{
		Text: "testvalue",
	}
	hw := NewHtmlWriter()
	hw.NewRoot("html").
		AddTextElement("body").
		AddTemplatedHTML("This is a bunch of raw text<p>{{.Text}}</p>", pl).
		AddTextElement("span", func(hte HtmlTextElement) { hte.AppendText("blah") }).
		AddNodeElement("div", func(hne HtmlNodeElement) {
			hne.AddTextElement("span", func(hte HtmlTextElement) { hte.SetID("12345") })
		})
	sb := &strings.Builder{}
	n, err := hw.WriteTo(sb)
	assert.Nil(t, err)
	assert.Equal(t, int64(sb.Len()), n)
	assert.Equal(t, "<html><body></body>This is a bunch of raw text<p>testvalue</p><span>blah</span><div><span ID=\"12345\"></span></div></html>", sb.String())
}
