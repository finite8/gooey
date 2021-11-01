package core

import "github.com/ntaylor-barnett/gooey/register"

type Styling interface {
	GetClasses(ctx register.PageContext) []string
	GetStyles(ctx register.PageContext) map[string]string
	CanApply(ctx register.PageContext, c Component) error
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

func (s *Style) CanApply(ctx register.PageContext, c Component) error {
	return nil
}
