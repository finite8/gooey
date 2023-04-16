package register

import (
	"net/http"
	"strings"
)

// a basic page info reference
type Page interface {
	//	ID() string
	Name() string
	Handler(ctx PageContext, w http.ResponseWriter, r *http.Request) interface{}
}

type ComplexPage interface {
	Page
	OnHandlerAdded(reg Registerer)
}

// PageLayour provides a way to do standard rendering of the layout of the entire page. This should not render HTML and HEAD elements, this is up to the GOOEY engine
type PageLayout interface {
	Render(ctx PageContext, w http.ResponseWriter, r *http.Request, pageRenderer func(ctx PageContext, w http.ResponseWriter, r *http.Request))
	// RenderLeading(ctx PageContext, w io.Writer)
	// RenderTrailing(ctx PageContext, w io.Writer)
	OnHandlerAdded(reg Registerer)
}

type PageBehaviour interface {
	QueryBehaviour(ctx PageContext, b Behaviour) Behaviour
}

type PageStructure interface {
	Page() Page
	Title() string
	Weight() int
	Children() []PageStructure
}

type PageStructureCollection []PageStructure

func (s PageStructureCollection) Len() int {
	return len(s)
}
func (s PageStructureCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s PageStructureCollection) Less(i, j int) bool {
	if s[i].Weight() == s[j].Weight() {
		return strings.Compare(s[i].Title(), s[j].Title()) < 0
	}
	return s[i].Weight() < s[j].Weight()
}

// Resolvable is something that can be translated into a string
type Resolvable interface {
	Resolve(ctx PageContext, kind ResolutionKind) string
}

type ResolutionKind string

const (
	Resolution_CSSClass = ResolutionKind("CssClass")
)
