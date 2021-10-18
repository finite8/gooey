package packages

import (
	"context"
	"fmt"
	"time"

	"example.com/goboganui/pkg/fancy"
	"example.com/goboganui/pkg/register"
)

func init() {
	register.RegisterPage(register.RootPageId, "listPackages", NewListPackages())
}

type listPackages struct {
	fancy.ContainerPage
	counter int
}

func NewListPackages() *listPackages {
	lp := &listPackages{}
	lp.WithName("List Packages").WithComponent(fancy.NewTableComponent(func(pc register.PageContext) (interface{}, error) {
		lp.counter++
		return []ContactDetails{
			{Email: "test.com",
				Subject: fmt.Sprintf("test %v", lp.counter),
				Message: "blergh"},
			{Email: "blargh.com",
				Subject: "gfdgfdgfd",
				Message: "gfdgdfgdfgdfgfdgfdgfdfgd"},
		}, nil
	})).WithComponent(fancy.NewStreamComponent(func(c1 context.Context, c2 chan<- string) error {
		var counter = 0
		for {
			select {
			case <-c1.Done():
				return nil
			case <-time.After(time.Second * 3):
				counter++
			}
			c2 <- fmt.Sprintf("%v", counter)
		}
	}))
	return lp
}

type ContactDetails struct {
	Email   string
	Subject string
	Message string
}

// func (lp *listPackages) Name() string {
// 	return "List Packages"
// }
// func (lp *listPackages) Handler(w http.ResponseWriter, r *http.Request) {
// 	tmpl := template.Must(template.ParseFiles("forms.html"))
// 	if r.Method != http.MethodPost {
// 		tmpl.Execute(w, nil)
// 		return
// 	}

// 	details := ContactDetails{
// 		Email:   r.FormValue("email"),
// 		Subject: r.FormValue("subject"),
// 		Message: r.FormValue("message"),
// 	}

// 	// do something with details
// 	_ = details

// 	tmpl.Execute(w, struct{ Success bool }{true})
// }
