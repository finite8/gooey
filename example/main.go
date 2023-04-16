// forms.go
package main

import (
	"net/http"
	"time"

	_ "github.com/ntaylor-barnett/gooey/example/pages"
	_ "github.com/ntaylor-barnett/gooey/pkg/bootstrap"
	"github.com/ntaylor-barnett/gooey/register"
	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
)

func main() {
	err := register.Compile()
	if err != nil {
		logrus.Fatal(err)
	}
	register.RegisterHandlers()
	go func() {
		time.Sleep(time.Second)
		browser.OpenURL("http://localhost:8080")
	}()
	http.ListenAndServe(":8080", nil)
}
