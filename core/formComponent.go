package core

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/uuid"
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
	uniqueId string
	//dataGetter         func(register.PageContext) (T, error)
	fstruct            *FormStructure
	defaultValueGetter func(register.PageContext) T
	onFormSubmitted    func(register.PageContext, T)
	KeepValues         bool
}

type FormTemplate[T interface{}] struct {
	// The struct to be used
	DefaultValue T
	// A map that can either mappings between the field and the rule, or between the field and another map for nested structures.
	FieldRules map[string]interface{}
}

func (ft *FormTemplate[T]) GetDefaultValue() interface{} {
	return ft.DefaultValue
}

func (ft *FormTemplate[T]) GetFieldRules() map[string]interface{} {
	return ft.FieldRules
}

func (fc *FormComponent[T]) WithKeepValues(keep bool) *FormComponent[T] {
	fc.KeepValues = keep
	return fc
}

type iformTemplate interface {
	GetDefaultValue() interface{}
	GetFieldRules() map[string]interface{}
}

type FieldRule struct {
	Required    bool
	Min         float64
	Max         float64
	RegexString string
}

func NewForm[T interface{}](defaultValueGetter func(register.PageContext) T) (*FormComponent[T], error) {
	fc := &FormComponent[T]{
		defaultValueGetter: defaultValueGetter,
		uniqueId:           uuid.New().String(),
	}
	var templateValue T
	fs, err := CreateFormStructure(templateValue)
	if err != nil {
		return nil, errors.Wrap(err, "failed to establish form structure")
	}
	fc.fstruct = fs
	return fc, nil
}

func MustNewForm[T interface{}](defaultValueGetter func(register.PageContext) T) *FormComponent[T] {
	fc, err := NewForm(defaultValueGetter)
	if err != nil {
		panic(err)
	}
	return fc
}

func (fc *FormComponent[T]) WriteContent(ctx register.PageContext, w PageWriter) {
	defaultValue := fc.defaultValueGetter(ctx)
	// if err != nil {
	// 	WriteComponentError(ctx, fc, err, w)
	// 	return
	// }

	// now we have a form structure, we can render it.

	var (
		validationFailures map[string]interface{}
		origValues         map[string]interface{}
	)
	if v, found := ctx.RequestCache().GetValue(fmt.Sprintf("VAL%s", fc.uniqueId)); found {
		validationFailures = v.(map[string]interface{})
	} else {
		validationFailures = make(map[string]interface{})
	}
	if v, found := ctx.RequestCache().GetValue(fmt.Sprintf("ORIG%s", fc.uniqueId)); found {
		origValues = v.(map[string]interface{})
	} else {
		origValues = make(map[string]interface{})
	}
	io.WriteString(w, `<form action="" method="post">`)
	for _, item := range fc.fstruct.Inputs {
		t := NewTag("div", map[string]interface{}{
			"class": "form-group",
		}, func() (retarr []Renderable) {
			retarr = []Renderable{
				NewTag("label", map[string]interface{}{"for": item.FieldName}, item.Label),
			}
			inputTag := NewUnpairedTag("input", nil)
			retarr = append(retarr, inputTag)
			// we need to check for validation status
			verr := validationFailures[item.FieldName]
			attribs := map[string]interface{}{
				"type":  "text",
				"class": "form-control",
				"id":    item.FieldName,
				"name":  item.FieldName,
			}
			if verr != nil {
				// it is in an invalid state
				attribs["class"] = "form-control is-invalid"
				retarr = append(retarr, NewTag("div", map[string]interface{}{"class": "invalid-feedback"}, verr))
			}
			if oval, ok := origValues[item.FieldName]; ok {
				attribs["value"] = oval
			} else {
				// now we need to see if we have a default value.
				dv := item.ValueGetter(defaultValue)
				if dv != nil {
					// because fmt.Sprint does not work well with pointers. We need to reflect first to dereference it
					rf := reflect.ValueOf(dv)
					if !rf.IsZero() {
						if rf.Kind() == reflect.Ptr {
							dv = rf.Elem().Interface()
						}
						attribs["value"] = dv
					}
				}
			}
			inputTag.Attributes = attribs
			return

		})
		w.WriteElement(ctx, t)
		// 		io.WriteString(w, fmt.Sprintf(`<div class="form-group">
		// 	<label for="%s">%s</label>
		// 	<input type="text" class="form-control" id="%s" name="%s">
		// </div>`, item.FieldName, item.Label, item.FieldName, item.FieldName))

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

func (fc *FormComponent[T]) HandlePost(ctx register.PageContext, r *http.Request) PostHandlerResult {
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
		validationErrors := map[string]interface{}{}
		origValues := map[string]interface{}{}
		for _, setValues := range fields {
			splt := strings.SplitN(setValues, "=", 2)
			tField := vmap[splt[0]]
			newVal := html.UnescapeString(splt[1])
			origValues[tField.FieldName] = newVal
			if tField.Validate != nil {
				valErr := tField.Validate(newVal)
				if valErr != "" {
					validationErrors[tField.FieldName] = valErr
					continue
				}
			}
			tField.ValueSetter(&outVal, newVal)
		}

		if len(validationErrors) == 0 {
			if fc.KeepValues {
				ctx.RequestCache().SetValue(fmt.Sprintf("ORIG%s", fc.uniqueId), origValues)
			}
			fc.onFormSubmitted(ctx, outVal)
		} else {
			ctx.RequestCache().SetValue(fmt.Sprintf("ORIG%s", fc.uniqueId), origValues)
			ctx.RequestCache().SetValue(fmt.Sprintf("VAL%s", fc.uniqueId), validationErrors)
		}
		return PostHandlerResult{
			IsHandled: true,
		}
	}
	return PostHandlerResult{}
}

func (fc *FormComponent[T]) OnRegister(rootCtx register.Registerer) {

}

// func (fc *FormComponent[T]) OnRegister(rootCtx register.Registerer) {
// 	formHandlePage := register.NewAPIPage("formsubmit", func(ctx register.PageContext, w http.ResponseWriter, r *http.Request) interface{} {
// 		switch r.Method {
// 		case http.MethodPost:
// 			if fc.onFormSubmitted != nil {
// 				data, _ := io.ReadAll(r.Body)
// 				fields := strings.Split(string(data), "&")
// 				fstruct := fc.fstruct
// 				vmap := map[string]*FormField{}
// 				var outVal T
// 				for _, v_iter := range fstruct.Inputs {
// 					v := v_iter
// 					vmap[v.FieldName] = v
// 				}
// 				for _, setValues := range fields {
// 					splt := strings.SplitN(setValues, "=", 2)
// 					tField := vmap[splt[0]]
// 					newVal := html.UnescapeString(splt[1])
// 					tField.ValueSetter(&outVal, newVal)
// 				}
// 				fc.onFormSubmitted(ctx, outVal)
// 			}
// 			return rootCtx.ThisPage()

// 		default:
// 			w.WriteHeader(405)
// 		}
// 		return nil
// 	})
// 	rootCtx.RegisterPrivateSubPage("formsubmit", formHandlePage)
// 	fc.formSubmitPage = formHandlePage
// }

func CreateFormStructure(base interface{}) (*FormStructure, error) {
	switch vt := base.(type) {
	case *FormStructure:
		return vt, nil
	case FormStructure:
		return &vt, nil
	}
	// if we get here, we are going to have to be a bit smarter and attempt to reflect it.
	var (
		rv         reflect.Value
		fieldRules map[string]interface{}
	)
	if formTemplate, ok := base.(iformTemplate); ok {
		templateValue := formTemplate.GetDefaultValue()
		rv = reflect.ValueOf(templateValue)
		fieldRules = formTemplate.GetFieldRules()
	} else {
		rv = reflect.ValueOf(base)
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		return reflectFormStructure(rv, fieldRules), nil
	}
	return nil, fmt.Errorf("don't know how to turn %T into a form", base)
}

func inferFieldRule(f reflect.StructField) *FieldRule {
	fieldrule := f.Tag.Get("gooey")
	if len(fieldrule) == 0 {
		return nil
	}
	var rule FieldRule
	for _, rv := range strings.Split(fieldrule, ",") {
		parts := strings.SplitN(rv, "=", 2)
		switch parts[0] {
		case "required":
			rule.Required = true
		case "min":
			if fv, e := strconv.ParseFloat(parts[1], 64); e != nil {
				panic(fmt.Errorf("min value for %s is invalid: %v", f.Name, e))
			} else {
				rule.Min = fv
			}
		case "max":
			if fv, e := strconv.ParseFloat(parts[1], 64); e != nil {
				panic(fmt.Errorf("max value for %s is invalid: %v", f.Name, e))
			} else {
				rule.Max = fv
			}
		case "regex":
			rString, e := url.QueryUnescape(parts[1])
			if e != nil {
				panic(fmt.Errorf("regex value for %s is invalid: %v", f.Name, e))
			}
			_, e = regexp.Compile(rString)
			if e != nil {
				panic(fmt.Errorf("regex value for %s could not be parsed: %v", f.Name, e))
			}
			rule.RegexString = rString

		}
	}
	return &rule
}

func reflectFormStructure(sv reflect.Value, fmappings map[string]interface{}) *FormStructure {
	newForm := &FormStructure{
		Title: sv.Type().Name(),
	}
	sinfo := sv.Type()
	for ix := 0; ix < sv.NumField(); ix++ {
		val := sv.Field(ix)
		info := sinfo.Field(ix)
		st := info.Type
		var frule *FieldRule
		// first part, lets try to see if we were given an explicit field rule to figure out.
		if fmappings != nil {
			if i, ok := fmappings[info.Name]; ok {
				// next, we need to make sure it is a value we support
				if fr, ok := i.(FieldRule); ok {
					frule = &fr
				}
			}
		}
		if frule == nil {
			frule = inferFieldRule(info)
		}
		var isNillable bool
		if st.Kind() == reflect.Ptr {
			isNillable = true
			st = st.Elem()
		}

		ff := &FormField{
			Label:        info.Name,
			FieldName:    info.Name,
			DefaultValue: val.Interface(),
		}
		fIx := ix
		ff.ValueGetter = func(i interface{}) interface{} {
			rVal := reflect.ValueOf(i)
			fld := rVal.Field(fIx)
			return fld.Interface()
		}
		if frule != nil {
			ff.Rule = *frule
		}

		switch st.Kind() {
		case reflect.String:
			ff.ValueType = StringType
			ff.ValueSetter = func(destStruct interface{}, value string) error {
				rVal := reflect.ValueOf(destStruct).Elem()
				fld := rVal.Field(fIx)
				if isNillable {
					rvts := reflect.ValueOf(&value)

					fld.Set(rvts)
				} else {
					fld.SetString(value)
				}
				return nil
			}
			ff.Validate = func(s string) string {
				if ff.Rule.Min != 0 && len(s) < int(ff.Rule.Min) {
					return fmt.Sprintf("requires a minimum of %d characters", int(ff.Rule.Min))
				}
				if ff.Rule.Max != 0 && len(s) > int(ff.Rule.Max) {
					return fmt.Sprintf("cannot be greather than %d characters", int(ff.Rule.Max))
				}
				if ff.Rule.Required {
					if strings.TrimSpace(s) == "" {
						return "required"
					}
				}
				if ff.Rule.RegexString != "" {
					m, err := regexp.Match(ff.Rule.RegexString, []byte(s))
					if err != nil {
						return err.Error()
					}
					if !m {
						return fmt.Sprintf("does not match the regex pattern: %s", ff.Rule.RegexString)
					}
				}
				return ""
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			ff.ValueType = IntType

			ff.ValueSetter = func(destStruct interface{}, value string) error {
				var (
					intval int64
					e      error
				)
				if value != "" {
					intval, e = strconv.ParseInt(value, 10, 64)
					if e != nil {
						return errors.New("not an int")
					}
				} else {
					intval = 0
				}
				rVal := reflect.ValueOf(destStruct).Elem()
				fld := rVal.Field(fIx)

				fld.SetInt(intval)
				return nil
			}
			ff.Validate = func(s string) string {
				if ff.Rule.Required {
					if strings.TrimSpace(s) == "" {
						return "required"
					}
				}
				intval, e := strconv.ParseInt(s, 10, 64)
				if e != nil {
					return "couldn't parse the given value as an int"
				}
				if ff.Rule.Min != 0 && intval < int64(ff.Rule.Min) {
					return fmt.Sprintf("cannot be less than %d", int64(ff.Rule.Min))
				}
				if ff.Rule.Max != 0 && intval > int64(ff.Rule.Max) {
					return fmt.Sprintf("cannot be greather than %d", int64(ff.Rule.Max))
				}

				if ff.Rule.RegexString != "" {
					m, err := regexp.Match(ff.Rule.RegexString, []byte(s))
					if err != nil {
						return err.Error()
					}
					if !m {
						return fmt.Sprintf("does not match the regex pattern: %s", ff.Rule.RegexString)
					}
				}
				return ""
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
	ValueGetter  func(interface{}) interface{}
	Validate     func(string) string
	Rule         FieldRule
}

type ReflectedValueSetter func(destStruct interface{}, value string) error

type FieldValueType byte

const (
	StructType = FieldValueType(iota)
	StringType
	IntType
)
