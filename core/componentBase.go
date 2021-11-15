package core

import "github.com/ntaylor-barnett/gooey/register"

// all components should share this
type ComponentBase struct {
	*register.AttibutingElement
	Style Styling
}

func (cb *ComponentBase) GetAttributes(ctx register.PageContext) string {
	m, _ := cb.AttibutingElement.GetAttributes(ctx)
	if cb.Style != nil {
		cb.Style.Apply(ctx, &m)
	}
	return register.MapToAttributes(m)
}
