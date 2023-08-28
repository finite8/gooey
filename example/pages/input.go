package pages

import (
	"encoding/json"
	"fmt"

	"github.com/finite8/gooey/core"
	"github.com/finite8/gooey/register"
)

var _ = register.RegisterPage(register.RootPageId, "inputPage", NewInput())

type inputPage struct {
	core.ContainerPage
	results []TestStruct
}

func NewInput() *inputPage {
	ip := &inputPage{}
	ip.WithName("Input Form").WithComponent(core.MustNewForm(func(pc register.PageContext) TestStruct {
		return TestStruct{
			Name: "default name",
		}
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
	Name       string `gooey:"min=2,max=10"`
	Address    string
	Religion   *string
	PretendAge *int32
	RealAge    int
	Sub        *SubStruct
}

type SubStruct struct {
	SubField     string
	NestedNested *AnotherSub
}

type AnotherSub struct {
	TestChildValue *string
	SomeNumber     uint64
}
