package core

import (
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/ntaylor-barnett/gooey/register"
)

// objectComponent is a basic representation of an object (something that has key-value pairs).
// this is designed for the representation of singletons and not slices/arrays. It will however create
// expandable boxes for child objects up to a specified depth.

type ObjectComponent struct {
	ComponentBase
	dataGetter func(register.PageContext) (interface{}, error)
}

func NewObjectComponent(f func(register.PageContext) (interface{}, error)) Component {
	return &ObjectComponent{
		dataGetter: f,
	}
}

type FieldValue struct {
	Label string
	Value interface{}
}

func getRenderableValues(item interface{}) (retarr []FieldValue, err error) {

	rv := reflect.ValueOf(item)
	if rv.Kind() == reflect.Map {
		if rv.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("cannot render %T as its key is not a string", item)
		}
		// we need to use special logic for maps that is actually easier
		var keys []string
		for _, k := range rv.MapKeys() {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		for _, k := range keys {
			key := k
			val := rv.MapIndex(reflect.ValueOf(k))
			retarr = append(retarr, FieldValue{
				Label: key,
				Value: val.Interface(),
			})
		}

	} else {
		isNill := false
		if rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				isNill = true
			}
			rv = rv.Elem()
		}
		rt := rv.Type()

		for ix := 0; ix < rt.NumField(); ix++ {
			field := rt.Field(ix)
			var val interface{}
			if !isNill {
				rfv := rv.Field(ix)
				if rfv.Kind() == reflect.Pointer {
					if rfv.IsNil() {
						val = nil
					} else {
						val = rfv.Elem().Interface()
					}
				} else {
					val = rfv.Interface()
				}
			}
			retarr = append(retarr, FieldValue{
				Label: field.Name,
				Value: val,
			})

		}
	}
	return

}

func (oc *ObjectComponent) OnRegister(ctx register.Registerer) {

}

func (oc *ObjectComponent) Write(ctx register.PageContext, w PageWriter) {
	data, err := oc.dataGetter(ctx)
	if err != nil {
		// we need to handle this somehow
		WriteComponentError(ctx, oc, err, w)
		return
	}
	fields, err := getRenderableValues(data)
	if err != nil {
		// we need to handle this somehow
		WriteComponentError(ctx, oc, err, w)
		return
	}
	// lets write the table parts
	io.WriteString(w, `<table class="GOOEY_table"><tr>`)
	for _, row := range fields {
		io.WriteString(w, `<tr>`)
		WriteElements(ctx, "<th>", "</th>", w, row.Label)
		WriteElements(ctx, "<td>", "</td>", w, makeRenderableInternal(row.Value, oc.currstate.opt, oc.currstate.dept))
		io.WriteString(w, `</tr>`)
	}
	io.WriteString(w, `</table>`)

}
