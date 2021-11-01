package register

import (
	"io"
	"net/http"
)

// a basic page info reference
type Page interface {
	//	ID() string
	Name() string
	Handler(ctx PageContext, w http.ResponseWriter, r *http.Request)
}

type ComplexPage interface {
	Page
	OnHandlerAdded(reg Registerer)
}

// PageLayour provides a way to do standard rendring of the layout of the entire page.
type PageLayout interface {
	RenderLeading(ctx PageContext, w io.Writer)
	RenderTrailing(ctx PageContext, w io.Writer)
	OnHandlerAdded(reg Registerer)
}

type PageBehaviour interface {
	QueryBehaviour(ctx PageContext, b Behaviour) Behaviour
}




