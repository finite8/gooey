package core

import (
	"fmt"
	"io"
	"reflect"

	"github.com/finite8/gooey/register"
)

type ListComponent struct {
	ComponentBase
	dataGetter func(register.PageContext) (interface{}, error)
}

func NewListComponent(f func(register.PageContext) (interface{}, error)) *ListComponent {
	return &ListComponent{
		dataGetter: f,
	}
}

func ArrayToRenderableArray(arrayOfValues interface{}) []Renderable {
	var retArr []Renderable
	rv := reflect.ValueOf(arrayOfValues)

	// now lets do the rows
	for ix := 0; ix < rv.Len(); ix++ {
		currItem := rv.Index(ix)
		if currItem.Kind() == reflect.Ptr {
			currItem = currItem.Elem()
		}
		currVal := currItem.Interface()
		if currVal == nil {
			retArr = append(retArr, nil)
			continue
		}
		if rndr, ok := currVal.(Renderable); ok {
			retArr = append(retArr, rndr)
			continue
		}
		// else, we don't REALLY know what we have here, so lets just turn it into text

		rndr_val := MakeRenderablePrimitive(currVal) // NewTextPrimitve(fmt.Sprintf("%v", currVal))
		retArr = append(retArr, rndr_val)

	}
	return retArr

}

func (lc *ListComponent) Write(ctx register.PageContext, w PageWriter) {
	data, err := lc.dataGetter(ctx)
	if err != nil {
		// we need to handle this somehow
		WriteComponentError(ctx, lc, err, w)
		return
	}
	var (
		renderList []Renderable
		ok         bool
	)
	if renderList, ok = data.([]Renderable); !ok {
		rv := reflect.ValueOf(data)
		rt := rv.Type()
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
			rv = rv.Elem()
		}
		// now we should have the right element
		switch rt.Kind() {
		case reflect.Array, reflect.Slice:
			renderList = ArrayToRenderableArray(data)
		default:
			WriteComponentError(ctx, lc, fmt.Errorf("%t cannot be represented as a renderable list", data), w)
			return
		}

	}

	// lets write the list parts
	io.WriteString(w, `<ul>`)
	for _, item := range renderList {
		io.WriteString(w, `<li>`)
		item.Write(ctx, w)
		io.WriteString(w, `</li>`)
	}
	io.WriteString(w, `</ul>`)

}

func (lc *ListComponent) OnRegister(ctx register.Registerer) {

}
