package core

import (
	"strings"

	"github.com/finite8/gooey/register"
)

type Styling interface {
	GetClasses(ctx register.PageContext) []string
	GetStyles(ctx register.PageContext) map[string]string
	Apply(ctx register.PageContext, attrMap *map[string]interface{}) error
}

type Style struct {
	styles  map[string]string
	classes []string
}

func NewClassStyling(classes ...string) *Style {
	return &Style{
		classes: classes,
	}
}

var _ Styling = (*Style)(nil)

func (s *Style) GetClasses(ctx register.PageContext) []string {
	return s.classes
}
func (s *Style) GetStyles(ctx register.PageContext) map[string]string {
	return s.styles
}

func (s *Style) Apply(ctx register.PageContext, attrMap *map[string]interface{}) error {
	if len(s.classes) > 0 {
		(*attrMap)["class"] = strings.Join(s.classes, " ")
	}
	return nil
}
