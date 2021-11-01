package register

import (
	"net/http"
)

type APIPage struct {
	name   string
	action func(ctx PageContext, w http.ResponseWriter, r *http.Request)
}

func (ap *APIPage) Name() string {
	return ap.name
}
func (ap *APIPage) Handler(ctx PageContext, w http.ResponseWriter, r *http.Request) {
	ap.action(ctx, w, r)
}

func (ap *APIPage) QueryBehaviour(ctx PageContext, b Behaviour) Behaviour {
	return b.WithRenderLayout(false)
}

func NewAPIPage(name string, f func(ctx PageContext, w http.ResponseWriter, r *http.Request)) *APIPage {
	return &APIPage{
		name:   name,
		action: f,
	}
}
