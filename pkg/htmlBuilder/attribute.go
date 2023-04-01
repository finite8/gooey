package htmlbuilder

import (
	"fmt"
	"strings"
)

type HtmlAttributeValue interface {
	GetValue() string                          // should be the text that is put into HTML
	ApplyValue(interface{}) HtmlAttributeValue // Apply the given value in a way that makes sense for the attribute.
	IsSet() bool                               // return true if it has a valid value and can be rendered. Return false otherwise
	MergeInto(HtmlAttributeValue)              // Copy the values from THIS attribute into the given one
}

// StringAttribute is the simplest type of attribute. It is just a string value. It cannot be appended, so applying a new value replaces its old one
type StringAttribute struct {
	Value string
}

func (sa *StringAttribute) GetValue() string {
	return sa.Value
}
func (sa *StringAttribute) ApplyValue(v interface{}) HtmlAttributeValue {
	switch val := v.(type) {
	case string:
		sa.Value = val
	case *string:
		sa.Value = *val
	default:
		sa.Value = fmt.Sprintf("%v", val)
	}
	return sa
}
func (sa *StringAttribute) IsSet() bool {
	return sa.Value != ""
}
func (sa *StringAttribute) MergeInto(a HtmlAttributeValue) {
	target := a.(*StringAttribute)
	target.Value = sa.Value

}

// ClassAttribute is formatted for class values. It can have multiple classes, and applying a new one appends it.
type ClassAttribute struct {
	Classes []string
}

func (ca *ClassAttribute) GetValue() string {
	return strings.Join(ca.Classes, " ")
}

func (ca *ClassAttribute) ApplyValue(v interface{}) HtmlAttributeValue {
	var classString string
	switch val := v.(type) {
	case string:
		classString = val
	case *string:
		classString = *val
	default:
		classString = fmt.Sprintf("%v", val)
	}
	ca.Classes = append(ca.Classes, strings.TrimSpace(classString))
	return ca
}
func (ca *ClassAttribute) IsSet() bool {
	return len(ca.Classes) > 0
}

func (ca *ClassAttribute) MergeInto(a HtmlAttributeValue) {
	target := a.(*ClassAttribute)
	for _, v := range ca.Classes {
		if !stringArrayContains(target.Classes, v) {
			target.Classes = append(target.Classes, v)
		}
	}

}

func stringArrayContains(arr []string, f string) bool {
	for _, v := range arr {
		if v == f {
			return true
		}
	}
	return false
}
