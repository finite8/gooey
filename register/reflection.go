package register

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func unmarshallFormData(data map[string][]string, target interface{}) error {
	// first, we need to inspect the thing we are unmarshalling into

	switch t := target.(type) {
	case *map[string]interface{}:
		for k, v := range data {
			(*t)[k] = v
		}
		return nil
	}
	rtv := reflect.ValueOf(target)
	if rtv.Kind() != reflect.Ptr {
		// the target should ALWAYS be a pointer to the real thing. Ignore any alternatives
		return fmt.Errorf("expected a pointer, but got %t instead", target)
	}
	// so, lets try to process this
	realTargetValue := rtv.Elem()
	realTargetType := realTargetValue.Type()
	// this will do much more forgiving lookups on the map
	for ix := 0; ix < realTargetType.NumField(); ix++ {
		field := realTargetType.Field(ix)

		valArr, ok := findMapValue(data, field.Name)

		if ok {
			// now we need to try and set it:
			fieldVal := realTargetValue.Field(ix)
			realType := field.Type
			for realType.Kind() == reflect.Ptr {
				realType = realType.Elem()
			}
			setFieldValue(valArr, fieldVal, realType)
		}
	}

	return nil
}

func setFieldValue(srcVal []string, rv reflect.Value, rt reflect.Type) {
	ref := getReflectValue(srcVal, rv, rt)
	if ref != nil {
		rv.Set(*ref)
	}
}

func getReflectValue(srcVal []string, rv reflect.Value, rt reflect.Type) *reflect.Value {
	var newVal interface{}
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		// we have an array. We need to get to the arrays REAL value
		// this recusive technique works great because our code works for single values and arrays of values
		arrayType := rt.Elem()
		// now we need some kind of reflect value that will take our array element
		for _, sv := range srcVal {

			newRefVal := getReflectValue([]string{sv}, reflect.New(arrayType).Elem(), arrayType)
			rv.Set(reflect.Append(rv, *newRefVal))
		}
		return &rv
	case reflect.Int:
		if len(srcVal) != 1 {
			return nil
		}
		newInt, _ := strconv.ParseInt(srcVal[0], 10, 64)
		newVal = int(newInt)
	case reflect.Int64:
		if len(srcVal) != 1 {
			return nil
		}
		newInt, _ := strconv.ParseInt(srcVal[0], 10, 64)
		newVal = int64(newInt)
	case reflect.Int32:
		if len(srcVal) != 1 {
			return nil
		}
		newInt, _ := strconv.ParseInt(srcVal[0], 10, 32)
		newVal = int32(newInt)
	case reflect.String:
		if len(srcVal) != 1 {
			return nil
		}
		sHolder := string(srcVal[0])
		newVal = sHolder
	}
	var ref reflect.Value

	if rv.Kind() == reflect.Ptr {
		if rt.Kind() == reflect.String {
			// we have to treat string pointers a little differet, as the method in the "else" completely falls over with a pointer to a string.
			// (i think the string gets GCd up or something, I'm kind of at the trial-and-error phase of this part of the code, so this works)
			strVal := newVal.(string)
			ref = reflect.ValueOf(&strVal)
		} else {
			ref = reflect.NewAt(rv.Type(), unsafe.Pointer(&newVal)).Elem()
		}

	} else {
		ref = reflect.ValueOf(newVal)
	}
	return &ref
}

func findMapValue(data map[string][]string, key string) ([]string, bool) {
	for k, v := range data {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}
