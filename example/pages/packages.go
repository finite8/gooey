package pages

import (
	"context"
	"fmt"
	"time"

	"github.com/ntaylor-barnett/gooey/core"
	"github.com/ntaylor-barnett/gooey/register"
)

func init() {
	register.RegisterPage(register.RootPageId, "listPackages", NewListPackages())
}

type listPackages struct {
	core.ContainerPage
	counter int
}

func NewListPackages() *listPackages {
	lp := &listPackages{}
	lp.WithName("List Packages").WithComponent(core.NewTableComponent(func(pc register.PageContext) (interface{}, error) {
		lp.counter++
		return []ContactDetails{
			{Email: "test.com",
				Subject: fmt.Sprintf("test %v", lp.counter),
				Message: "blergh"},
			{Email: "blargh.com",
				Subject: "gfdgfdgfd",
				Message: "gfdgdfgdfgdfgfdgfdgfdfgd"},
		}, nil
	})).WithComponent(core.NewStreamComponent(func(c1 context.Context, c2 chan<- string) error {
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
