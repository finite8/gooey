package pages

import (
	"encoding/json"
	"fmt"

	"github.com/ntaylor-barnett/gooey/core"
	"github.com/ntaylor-barnett/gooey/register"
)

var _ = register.RegisterPage(register.RootPageId, "inputPage", NewInput())

type inputPage struct {
	core.ContainerPage
}

func NewInput() *inputPage {
	ip := &inputPage{}
	ip.WithName("Input Form").WithComponent(core.NewForm(func(pc register.PageContext) (TestStruct, error) {
		return TestStruct{
			Name: "default name",
		}, nil
	}).WithSubmitHandler(func(pc register.PageContext, ts TestStruct) {
		d, _ := json.Marshal(ts)
		fmt.Println(string(d))
	}))
	return ip
}

type TestStruct struct {
	Name    string
	Address string
	Age     int
}
