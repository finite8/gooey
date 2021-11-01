// forms.go
package main

import (
	"net/http"

	_ "github.com/ntaylor-barnett/gooey/example/pages"
	"github.com/ntaylor-barnett/gooey/register"
	"github.com/sirupsen/logrus"
)

func main() {
	err := register.Compile()
	if err != nil {
		logrus.Fatal(err)
	}
	register.RegisterHandlers()

	http.ListenAndServe(":8080", nil)
}
