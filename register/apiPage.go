package register

import (
	"net/http"
)

type APIPage struct {
	name   string
	action func(ctx PageContext, w http.ResponseWriter, r *http.Request) interface{}
}

func (ap *APIPage) Name() string {
	return ap.name
}
func (ap *APIPage) Handler(ctx PageContext, w http.ResponseWriter, r *http.Request) interface{} {
	return ap.action(ctx, w, r)
}

func (ap *APIPage) QueryBehaviour(ctx PageContext, b Behaviour) Behaviour {
	return b.WithRenderLayout(false).WithRenderHTML(false)
}

func NewAPIPage(name string, f func(ctx PageContext, w http.ResponseWriter, r *http.Request) interface{}) *APIPage {
	return &APIPage{
		name:   name,
		action: f,
	}
}
