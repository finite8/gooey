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
	results []TestStruct
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
		ip.results = append(ip.results, ts)
	}))
	ip.WithComponent(core.NewTableComponent(func(pc register.PageContext) (interface{}, error) {
		return ip.results, nil
	}))
	return ip
}

type TestStruct struct {
	Name    string `gooey:"min=2,max=10"`
	Address string
	Age     int
}
