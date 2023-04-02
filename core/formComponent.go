package core

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/ntaylor-barnett/gooey/register"
)

/*
	This needs to do the following:
	- Provide a form so a user can input values
	- Provide a mechanism so that the data can be returned and processed
	- Have that data be pushed to another on screen component
*/

// FormComponent
type FormComponent[T interface{}] struct {
	ComponentBase
	dataGetter      func(register.PageContext) (T, error)
	formSubmitPage  register.Page
	fstruct         *FormStructure
	defaultValue    T
	onFormSubmitted func(register.PageContext, T)
}

func NewForm[T interface{}](dataGetter func(register.PageContext) (T, error)) *FormComponent[T] {
	return &FormComponent[T]{
		dataGetter: dataGetter,
	}
}

func (fc *FormComponent[T]) WriteContent(ctx register.PageContext, w PageWriter) {
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
	fc.fstruct = fs
	// now we have a form structure, we can render it.
	io.WriteString(w, fmt.Sprintf(`<form action="%s" method="post">`, ctx.GetPageUrl(fc.formSubmitPage)))
	for _, item := range fs.Inputs {
		io.WriteString(w, fmt.Sprintf(`<div class="form-group">
	<label for="%s">%s</label>
	<input type="text" class="form-control" id="%s" name="%s">
</div>`, item.FieldName, item.Label, item.FieldName, item.FieldName))

	}
	io.WriteString(w, `<button type="submit" class="btn btn-primary">Submit</button>`)
	io.WriteString(w, `</form>`)
}

func (fc *FormComponent[T]) WithSubmitHandler(f func(register.PageContext, T)) *FormComponent[T] {
	if fc.onFormSubmitted != nil {
		panic("onFormSubmitted has already been bound")
	}
	fc.onFormSubmitted = f
	return fc
}

func (fc *FormComponent[T]) OnRegister(ctx register.Registerer) {
	formHandlePage := register.NewAPIPage("formsubmit", func(ctx register.PageContext, w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if fc.onFormSubmitted != nil {
				data, _ := io.ReadAll(r.Body)
				fields := strings.Split(string(data), "&")
				fstruct := fc.fstruct
				vmap := map[string]*FormField{}
				var outVal T
				for _, v_iter := range fstruct.Inputs {
					v := v_iter
					vmap[v.FieldName] = v
				}
				for _, setValues := range fields {
					splt := strings.SplitN(setValues, "=", 2)
					tField := vmap[splt[0]]
					newVal := html.UnescapeString(splt[1])
					tField.ValueSetter(&outVal, newVal)
				}
				fc.onFormSubmitted(ctx, outVal)
			}

		default:
			w.WriteHeader(405)
		}
	})
	ctx.RegisterPrivateSubPage("formsubmit", formHandlePage)
	fc.formSubmitPage = formHandlePage
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
		fIx := ix
		switch st.Kind() {
		case reflect.String:
			ff.ValueType = StringType
			ff.ValueSetter = func(destStruct interface{}, value string) error {
				rVal := reflect.ValueOf(destStruct).Elem()
				fld := rVal.Field(fIx)
				fld.SetString(value)
				return nil
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			ff.ValueType = IntType

			ff.ValueSetter = func(destStruct interface{}, value string) error {
				intval, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return errors.New("not an int")
				}
				rVal := reflect.ValueOf(destStruct).Elem()
				fld := rVal.Field(fIx)

				fld.SetInt(intval)
				return nil
			}
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
	ValueSetter  ReflectedValueSetter
}

type ReflectedValueSetter func(destStruct interface{}, value string) error

type FieldValueType byte

const (
	StructType = FieldValueType(iota)
	StringType
	IntType
)
