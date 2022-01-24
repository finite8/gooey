package core

import (
	"fmt"
	"io"
	"reflect"

	"github.com/ntaylor-barnett/gooey/register"
)

/*
	This needs to do the following:
	- Provide a form so a user can input values
	- Provide a mechanism so that the data can be returned and processed
	- Have that data be pushed to another on screen component
*/

// FormComponent
type FormComponent struct {
	dataGetter func(register.PageContext) (interface{}, error)
}

func NewForm(dataGetter func(register.PageContext) (interface{}, error)) *FormComponent {
	return &FormComponent{
		dataGetter: dataGetter,
	}
}

func (fc *FormComponent) WriteContent(ctx register.PageContext, w io.Writer) {
	formData, err := fc.dataGetter(ctx)
	if err != nil {
		WriteComponentError(ctx, fc, err, w)
		return
	}
	fs, err := CreateFormStructure(formData)
	if err != nil {
		WriteComponentError(ctx, fc, err, w)
		return
	}
	// now we have a form structure, we can render it.
}

func (fc *FormComponent) OnRegister(ctx register.Registerer) {

}

func CreateFormStructure(base interface{}) (*FormStructure, error) {
	switch vt := base.(type) {
	case *FormStructure:
		return vt, nil
	case FormStructure:
		return &vt, nil
	}
	// if we get here, we are going to have to be a bit smarter and attempt to reflect it.
	rv := reflect.ValueOf(base)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		return reflectFormStructure(rv), nil
	}
	return nil, fmt.Errorf("don't know how to turn %T into a form", base)
}

func reflectFormStructure(sv reflect.Value) *FormStructure {
	newForm := &FormStructure{
		Title: sv.Type().Name(),
	}
	sinfo := sv.Type()
	for ix := 0; ix < sv.NumField(); ix++ {
		val := sv.Field(ix)
		info := sinfo.Field(ix)
		st := info.Type
		//var isNillable bool
		if st.Kind() == reflect.Ptr {
			//isNillable = true
			st = st.Elem()
		}
		ff := &FormField{
			Label:        info.Name,
			FieldName:    info.Name,
			DefaultValue: val.Interface(),
		}
		switch st.Kind() {
		case reflect.String:
			ff.ValueType = StringType
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64:
			ff.ValueType = IntType
		default:
			// the type isn't supported
		}

		newForm.Inputs = append(newForm.Inputs, ff)

	}

	return newForm
}

type FormStructure struct {
	Title  string
	Inputs []*FormField
}

type FormField struct {
	// Label is the text that describes it
	Label string
	// FieldName is the means to look up the struct that populated it
	FieldName    string
	ValueType    FieldValueType
	DefaultValue interface{}
}

type FieldValueType byte

const (
	StructType = FieldValueType(iota)
	StringType
	IntType
)
