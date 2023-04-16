package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRenderableValues_Map(t *testing.T) {
	oneString := "one"
	var fourString *string
	arr, err := getRenderableValues(map[string]interface{}{
		"5": "Five",
		"2": 2,
		"3": 3.0,
		"1": &oneString,
		"4": fourString,
	})
	assert.Nil(t, err)
	assert.Equal(t, arr, []FieldValue{
		{Label: "1", Value: &oneString},
		{Label: "2", Value: 2},
		{Label: "3", Value: 3.0},
		{Label: "4", Value: fourString},
		{Label: "5", Value: "Five"},
	})
}

func TestGetRenderableValues_Struct(t *testing.T) {

	type TestStruct struct {
		This   string
		Will   string
		Not    string
		Be     string
		Sorted string
	}
	arr, err := getRenderableValues(TestStruct{"Never", "Gonna", "Give", "You", "Up"})
	assert.Nil(t, err)
	assert.Equal(t, arr, []FieldValue{
		{Label: "This", Value: "Never"},
		{Label: "Will", Value: "Gonna"},
		{Label: "Not", Value: "Give"},
		{Label: "Be", Value: "You"},
		{Label: "Sorted", Value: "Up"},
	})
}
