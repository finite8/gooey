package register

import "net/http"

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
