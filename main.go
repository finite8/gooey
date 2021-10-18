// forms.go
package main

import (
	"net/http"

	_ "example.com/goboganui/pages/packages"
	"example.com/goboganui/pkg/register"
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
