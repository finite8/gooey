// forms.go
package main

import (
	"net/http"

	_ "github.com/ntaylor-barnett/gooey/pages/packages"
	"github.com/ntaylor-barnett/gooey/pkg/register"
	"github.com/sirupsen/logrus"
)

func main() {
	// tmpl := template.Must(template.ParseFiles("forms.html"))

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
	// })
	err := register.Compile()
	if err != nil {
		logrus.Fatal(err)
	}
	register.RegisterHandlers()

	http.ListenAndServe(":8080", nil)
}
