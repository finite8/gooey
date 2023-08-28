package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/finite8/gooey/register"
	"github.com/google/uuid"
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

func buildFormField(item *FormField, defaultValue interface{}, origValues, validationFailures PathedMap[string]) (retarr []Renderable) {
	retarr = []Renderable{
		NewTag("label", map[string]interface{}{
			"for":   item.FieldName,
			"class": "GOOEY_formlabel"}, item.Label),
	}
	inputTag := NewUnpairedTag("input", nil)
	retarr = append(retarr, inputTag)
	// we need to check for validation status
	verr, vErrExists := validationFailures.Get(item.Path)
	attribs := map[string]interface{}{
		"type":  "text",
		"class": "GOOEY_forminput",
		"id":    item.Path,
		"name":  item.Path,
	}
	if vErrExists {
		// it is in an invalid state
		attribs["class"] = "GOOEY_forminput is-invalid"
		retarr = append(retarr, NewTag("div", map[string]interface{}{"class": "invalid-feedback"}, verr))
	}
	if oval, ok := origValues.Get(item.Path); ok {
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
}

func getSubMap(inMap map[string]interface{}, key string) map[string]interface{} {
	if v, ok := inMap[key]; ok {
		if v, ok := v.(map[string]interface{}); ok {
			return v
		}
	}
	return nil
}

func buildFormElements(f *FormStructure, defaultValue interface{}, origValues, validationFailures PathedMap[string]) (retarr []Renderable) {
	for _, item := range f.Inputs {
		t := NewTag("div", map[string]interface{}{
			"class": "GOOEY_formgroup",
		}, func() (retarr []Renderable) {
			if item.SubStructure == nil {
				arr := buildFormField(item, defaultValue, origValues, validationFailures)
				return arr

			} else {
				dv := item.ValueGetter(defaultValue)
				if dv == nil || reflect.ValueOf(dv).IsZero() {
					dv = item.DefaultValue
				}
				arr := buildFormElements(item.SubStructure, dv, origValues, validationFailures)
				retArr := []Renderable{
					NewTag("label", map[string]interface{}{
						"for":   item.Path,
						"class": "GOOEY_formlabel"}, item.Label),
				}
				retArr = append(retArr, arr...)

				return []Renderable{
					NewTag("div", map[string]interface{}{
						"class": "GOOEY_formgroup",
					}, retArr),
				}
			}

		})
		retarr = append(retarr, t)
		// 		io.WriteString(w, fmt.Sprintf(`<div class="form-group">
		// 	<label for="%s">%s</label>
		// 	<input type="text" class="form-control" id="%s" name="%s">
		// </div>`, item.FieldName, item.Label, item.FieldName, item.FieldName))

	}
	return
}

func (fc *FormComponent[T]) Write(ctx register.PageContext, w PageWriter) {
	defaultValue := fc.defaultValueGetter(ctx)
	// if err != nil {
	// 	WriteComponentError(ctx, fc, err, w)
	// 	return
	// }

	// now we have a form structure, we can render it.

	var (
		validationFailures PathedMap[string]
		origValues         PathedMap[string]
	)
	if v, found := ctx.RequestCache().GetValue(fmt.Sprintf("VAL%s", fc.uniqueId)); found {
		validationFailures = v.(PathedMap[string])
	} else {
		validationFailures = make(PathedMap[string])
	}
	if v, found := ctx.RequestCache().GetValue(fmt.Sprintf("ORIG%s", fc.uniqueId)); found {
		origValues = v.(PathedMap[string])
	} else {
		origValues = make(PathedMap[string])
	}
	io.WriteString(w, `<form action="" method="post">`)
	{
		formElements := buildFormElements(fc.fstruct, defaultValue, origValues, validationFailures)
		w.WriteElement(ctx, formElements)
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

type PathedMap[T interface{}] map[string]interface{}

func (pm PathedMap[T]) Get(path string) (T, bool) {
	var defaultValue T
	parts := strings.SplitN(path, ".", 2)
	mv, ok := pm[parts[0]]
	if !ok {
		return defaultValue, false
	}
	if len(parts) == 2 {
		subMap, ok := mv.(PathedMap[T])
		if !ok {
			return defaultValue, false // this could be a panic but we will treat it as not found
		}
		return subMap.Get(parts[1])
	}
	if retVal, ok := mv.(T); ok {
		return retVal, true
	} else {
		return defaultValue, false
	}
}

func (pm PathedMap[T]) Set(path string, value T) {
	parts := strings.SplitN(path, ".", 2)
	mv, ok := pm[parts[0]]
	if !ok {
		// great, it doesn't exist. Lets set it
		if len(parts) == 2 {
			// it has more than one path. We have to continue down the chain.
			subMap := make(PathedMap[T])
			pm[parts[0]] = subMap
			subMap.Set(parts[1], value)
			return // handled. return.
		}
	} else {
		// the value already exists. Maybe uh oh
		if len(parts) == 2 {
			// we now need to test to see if this is a complex path or not
			subMap, ok := mv.(PathedMap[T])
			if !ok {
				subMap := make(PathedMap[T])
				pm[parts[0]] = subMap // if it was not a submap, then we overwrite it with a submap
			}
			subMap.Set(parts[1], value)
			return
		}
	}
	pm[parts[0]] = value
}

func (fc *FormComponent[T]) HandlePost(ctx register.PageContext, r *http.Request) PostHandlerResult {
	if fc.onFormSubmitted != nil {
		if len(r.Form) == 0 {
			r.ParseForm()
		}
		//data, _ := io.ReadAll(r.Body)
		//fields := strings.Split(string(data), "&")
		fstruct := fc.fstruct

		var outVal T
		vmap := fstruct.GetMap()
		validationErrors := make(PathedMap[string])
		origValues := make(PathedMap[string])
		for key, value := range r.Form {
			newVal := value[0]
			tField, err := vmap.GetField(key)
			if err != nil {
				return PostHandlerResult{
					IsHandled:      false,
					HaltProcessing: false,
					Error:          err,
				}
			}
			origValues.Set(tField.Path, newVal)
			if tField.Validate != nil {
				valErr := tField.Validate(newVal)
				if valErr != "" {
					validationErrors.Set(tField.Path, valErr)
					continue
				}
			}
			err = vmap.SetFieldValue(key, &outVal, newVal)
			if err != nil {
				return PostHandlerResult{
					IsHandled:      false,
					HaltProcessing: true,
					Error:          err,
				}
			}

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
		return reflectFormStructure("", rv, fieldRules)
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

func reflectFormStructure(prefix string, sv reflect.Value, fmappings map[string]interface{}) (*FormStructure, error) {
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
		var defaultValue interface{}
		if !val.IsZero() {
			if isNillable {
				defaultValue = val.Elem().Interface()
			} else {
				defaultValue = val.Interface()
			}
		} else {

			defaultValue = reflect.New(st).Elem().Interface()
		}
		ff := &FormField{
			Label:        info.Name,
			FieldName:    info.Name,
			DefaultValue: defaultValue,
		}
		if prefix == "" {
			ff.Path = info.Name
		} else {
			ff.Path = fmt.Sprintf("%s.%s", prefix, info.Name)
		}
		fIx := ix
		ff.ValueGetter = func(i interface{}) interface{} {
			rVal := reflect.ValueOf(i)

			for rVal.Kind() == reflect.Pointer {
				if rVal.IsNil() {
					return nil
				}
				rVal = rVal.Elem()

			}
			// if rVal.IsZero() {
			// 	return nil
			// }
			fld := rVal.Field(fIx)
			if fld.Kind() == reflect.Pointer {
				if fld.IsNil() {
					return nil
				}
			}
			// if fld.IsZero() {
			// 	if fld.Kind() == reflect.Pointer {
			// 		return nil
			// 	}
			// }
			return fld.Interface()
		}
		if frule != nil {
			ff.Rule = *frule
		}

		switch st.Kind() {
		case reflect.String:
			ff.ValueType = StringType
			ff.ValueSetter = func(destStruct interface{}, value string) error {
				rVal := reflect.ValueOf(destStruct)
				for rVal.Kind() == reflect.Pointer {
					rVal = rVal.Elem()
				}

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
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			ff.ValueType = IntType
			iType := st.Kind()
			ff.ValueSetter = func(destStruct interface{}, value string) error {
				var (
					intval uint64
					e      error
				)
				if value != "" {
					intval, e = strconv.ParseUint(value, 10, 64)
					if e != nil {
						return errors.New("not an uint")
					}
				} else {
					// there is no value, we should just return
					return nil
					// intval = 0
				}
				rVal := reflect.ValueOf(destStruct).Elem()
				fld := rVal.Field(fIx)
				if isNillable {
					var rvts reflect.Value
					// we are going to need to get this to the right type otherwise it won't work.
					switch iType {
					case reflect.Uint:
						v := uint(intval)
						rvts = reflect.ValueOf(&v)
					case reflect.Uint16:
						v := uint16(intval)
						rvts = reflect.ValueOf(&v)
					case reflect.Uint32:
						v := uint32(intval)
						rvts = reflect.ValueOf(&v)
					default:
						rvts = reflect.ValueOf(&intval)
					}
					fld.Set(rvts)
				} else {
					fld.SetUint(intval)
				}

				return nil
			}
			ff.Validate = func(s string) string {
				if strings.TrimSpace(s) == "" {
					if ff.Rule.Required {
						return "required"
					} else {
						return ""
					}
				}
				intval, e := strconv.ParseUint(s, 10, 64)
				if e != nil {
					return "couldn't parse the given value as an unsigned integer"
				}

				if ff.Rule.Min != 0 && intval < uint64(ff.Rule.Min) {
					return fmt.Sprintf("cannot be less than %d", uint64(ff.Rule.Min))
				}
				if ff.Rule.Max != 0 && intval > uint64(ff.Rule.Max) {
					return fmt.Sprintf("cannot be greather than %d", uint64(ff.Rule.Max))
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
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			ff.ValueType = IntType
			iType := st.Kind()
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
					// there is no value, we should just return
					return nil
				}
				rVal := reflect.ValueOf(destStruct).Elem()
				fld := rVal.Field(fIx)
				if isNillable {
					var rvts reflect.Value
					// we are going to need to get this to the right type otherwise it won't work.
					switch iType {
					case reflect.Int:
						v := int(intval)
						rvts = reflect.ValueOf(&v)
					case reflect.Int16:
						v := int16(intval)
						rvts = reflect.ValueOf(&v)
					case reflect.Int32:
						v := int32(intval)
						rvts = reflect.ValueOf(&v)
					default:
						rvts = reflect.ValueOf(&intval)
					}
					fld.Set(rvts)
				} else {
					fld.SetInt(intval)
				}

				return nil
			}
			ff.Validate = func(s string) string {

				if strings.TrimSpace(s) == "" {
					if ff.Rule.Required {
						return "required"
					} else {
						return ""
					}
				}
				intval, e := strconv.ParseInt(s, 10, 64)
				if e != nil {
					return "couldn't parse the given value as an integer"
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
		case reflect.Struct:
			if !isNillable {
				return nil, errors.Errorf("invalid template structure: Field %s of type %s must be a pointer", ff.Path, st.Name())
			}
			structVal := reflect.New(st).Elem()
			var subMappings map[string]interface{}
			if v, ok := fmappings[info.Name]; ok {
				if v, ok := v.(map[string]interface{}); ok {
					subMappings = v
				}
			}
			ss, err := reflectFormStructure(ff.Path, structVal, subMappings)
			if err != nil {
				return nil, err
			}
			ff.SubStructure = ss
		default:
			// the type isn't supported
		}

		newForm.Inputs = append(newForm.Inputs, ff)

	}

	return newForm, nil
}

type FormStructure struct {
	Title  string
	Inputs []*FormField
}

func (fs *FormStructure) GetMap() FormFieldMap {
	ffm := make(FormFieldMap)
	for _, i := range fs.Inputs {
		item := i
		ffm[i.FieldName] = item
	}
	return ffm
}

type FormField struct {
	// Label is the text that describes it
	Label string
	// FieldName is the means to look up the struct that populated it
	FieldName    string
	Path         string
	ValueType    FieldValueType
	DefaultValue interface{}
	ValueSetter  ReflectedValueSetter
	ValueGetter  func(interface{}) interface{}
	Validate     func(string) string
	Rule         FieldRule
	SubStructure *FormStructure
}

type FormFieldMap map[string]*FormField

func (ffm FormFieldMap) GetField(path string) (*FormField, error) {
	parts := strings.SplitN(path, ".", 2)
	ff, ok := ffm[parts[0]]
	if !ok {
		return nil, errors.Errorf("field %s was specified but not defined", parts[0])
	}
	if len(parts) == 2 {
		// there are two parts, so we have a sub-stucture.
		return ff.SubStructure.GetMap().GetField(parts[1])
	} else {
		return ff, nil
	}
}

func (ffm FormFieldMap) SetFieldValue(path string, destStruct interface{}, value string) error {
	parts := strings.SplitN(path, ".", 2)
	ff, ok := ffm[parts[0]]
	if !ok {
		return errors.Errorf("field %s was specified but not defined", parts[0])
	}
	if ff.SubStructure == nil {
		return ff.ValueSetter(destStruct, value)
	} else {
		// we have a sub structure. First we need to see if it needs to be initialized
		rv := ff.ValueGetter(destStruct)
		if rv == nil {
			// it is not set, so we need to initialize it.
			nv := reflect.New(reflect.ValueOf(ff.DefaultValue).Type())
			destReflect := reflect.ValueOf(destStruct)
			for destReflect.Kind() == reflect.Pointer {
				destReflect = destReflect.Elem()
			}
			destReflect.FieldByName(ff.FieldName).Set(nv)
			rv = nv.Interface()
		}
		return ff.SubStructure.GetMap().SetFieldValue(parts[1], rv, value)

	}
}

type ReflectedValueSetter func(destStruct interface{}, value string) error

type FieldValueType byte

const (
	StructType = FieldValueType(iota)
	StringType
	IntType
)
