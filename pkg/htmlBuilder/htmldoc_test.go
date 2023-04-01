package htmlbuilder

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlankHtmlDocument(t *testing.T) {
	doc := &htmldoc{}
	sb := &strings.Builder{}
	n, e := doc.Write(sb)
	assert.Nil(t, e)
	assert.Greater(t, n, 0)
	assert.Equal(t, sb.String(), "<HTML><HEAD/><TITLE/></HTML>")
}

func TestHtmlDocument(t *testing.T) {
	doc := &htmldoc{}
	doc.Head().Meta()["DC.identifier"] = NewAttributable().WithStringValues("content", "http://www.ietf.org/rfc/rfc1866.txt")
	doc.Body().AppendNode(NewHtmlContainer("div"))
	sb := &strings.Builder{}
	n, e := doc.Write(sb)
	assert.Nil(t, e)
	assert.Greater(t, n, 0)
	assert.Equal(t, sb.String(), `<HTML><HEAD><meta content="http://www.ietf.org/rfc/rfc1866.txt" name="DC.identifier"></HEAD><TITLE/><BODY><div/></BODY></HTML>`)
}
