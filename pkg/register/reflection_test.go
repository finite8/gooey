package register

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UnmarhsallMap(t *testing.T) {
	formData := map[string][]string{
		"first":  []string{"1"},
		"second": []string{"two, three"},
	}
	target := map[string]interface{}{}
	err := unmarshallFormData(formData, target)
	assert.Error(t, err)
	err = unmarshallFormData(formData, &target)
	assert.Nil(t, err)
	assert.Equal(t, target["first"], []string{"1"})
	assert.Equal(t, target["second"], []string{"two, three"})
}

func Test_UnmarhsallStruct(t *testing.T) {
	formData := map[string][]string{
		"IntegerVal":          []string{"1"},
		"IntegerArrayMany":    []string{"1", "2", "3"},
		"IntegerArrayOne":     []string{"1"},
		"IntegerPtrVal":       []string{"1"},
		"IntegerPtrArrayMany": []string{"1", "2", "3"},
		"StringVal":           []string{"first"},
		"StringArrayMany":     []string{"one", "two", "three"},
		"StringArrayOne":      []string{"one"},
		"StringPtr":           []string{"second"},
	}
	var target struct {
		IntegerVal          int
		IntegerArrayMany    []int
		IntegerArrayOne     []int
		IntegerPtrVal       *int
		IntegerPtrArrayMany []*int
		StringVal           string
		StringArrayMany     []string
		StringArrayOne      []string
		StringPtr           *string
	}
	err := unmarshallFormData(formData, &target)
	assert.Nil(t, err)
	assert.Equal(t, target.IntegerVal, 1)
	assert.Equal(t, target.IntegerArrayMany, []int{1, 2, 3})
	assert.Equal(t, target.IntegerArrayOne, []int{1})
	assert.Equal(t, target.StringVal, "first")
	assert.Equal(t, target.StringArrayMany, []string{"one", "two", "three"})
	assert.Equal(t, target.StringArrayOne, []string{"one"})
	assert.Equal(t, *target.StringPtr, "second")
}
