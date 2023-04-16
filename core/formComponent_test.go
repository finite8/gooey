package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathedMap(t *testing.T) {
	pm := make(PathedMap[string])
	pm.Set("root1", "1")
	pm.Set("root2", "2")
	pm.Set("nested.N1", "3")
	pm.Set("nested.N2", "4")
	assert.Equal(t, pm, PathedMap[string]{
		"root1": "1",
		"root2": "2",
		"nested": PathedMap[string]{
			"N1": "3",
			"N2": "4",
		},
	})
	{
		v, ok := pm.Get("root1")
		assert.Equal(t, v, "1")
		assert.True(t, ok)
	}
	{
		v, ok := pm.Get("nested.N1")
		assert.Equal(t, v, "3")
		assert.True(t, ok)
	}
	{
		v, ok := pm.Get("jkjlk")
		assert.Equal(t, v, "")
		assert.False(t, ok)
	}
	{
		v, ok := pm.Get("nested")
		assert.Equal(t, v, "")
		assert.False(t, ok)
	}
}

func TestReflectValid(t *testing.T) {
	type AnotherSub struct {
		TestChildValue *string
		SomeNumber     uint64
	}
	type SubStruct struct {
		SubField     string
		NestedNested *AnotherSub
	}
	type TestStruct struct {
		Name       string `gooey:"min=2,max=10"`
		Address    string
		Religion   *string
		PretendAge *int32
		RealAge    int
		Sub        *SubStruct
	}
	base := TestStruct{
		Name: "example",
	}
	fs, err := CreateFormStructure(&base)
	assert.NoError(t, err)
	_ = fs
}

func TestReflectInValid(t *testing.T) {
	{ // case 1: a field that has a struct as its type must be a pointer to that struct
		type SubStruct struct {
			SubField string
		}
		type TestStruct struct {
			Name string

			Sub SubStruct
		}
		base := TestStruct{
			Name: "example",
		}
		fs, err := CreateFormStructure(&base)
		assert.Contains(t, err.Error(), "must be a pointer")
		_ = fs
	}
}
