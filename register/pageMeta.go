package register

import (
	"fmt"
	"io"
	"strings"
)

// page meta is required for any and all pages. The system is designed to handle this as much as practical.

type AttibutingElement struct {
	Attributes map[string]interface{}
}

func (ae *AttibutingElement) GetAttributes(ctx PageContext) (map[string]interface{}, error) {
	if ae == nil {
		return map[string]interface{}{}, nil
	}
	return cloneMap(ae.Attributes), nil
}

func (ae *AttibutingElement) SetAttribute(attr string, val interface{}) {
	if ae.Attributes == nil {
		ae.Attributes = make(map[string]interface{})
	}
	ae.Attributes[strings.ToLower(attr)] = val
}

type PageElement struct {
	*AttibutingElement
	ElementName string
	Kind        ElementTagKind

	InnerText string
}

func (pe *PageElement) SetElementName(name string) {
	pe.ElementName = strings.ToLower(name)
}

type ElementTagKind byte

const (
	// <ElementName>InnerText</ElementName>
	ElementTag_Closing = 0
	// <ElementName>
	ElementTag_NoClose = 1
	// <ElementName />
	ElementTag_SelfClose = 2
)

type EditableElement interface {
	SetAttribute(attr string, val interface{})
	SetElementName(name string)
}

type RenderableElement interface {
	GetAttributes(ctx PageContext) (map[string]interface{}, error)
	GetElementName(ctx PageContext) string
	GetKind() ElementTagKind
	GetInnerText(ctx PageContext) string
}

func cloneMap(m map[string]interface{}) map[string]interface{} {
	retMap := make(map[string]interface{}, len(m))
	for k, v := range m {
		retMap[k] = v
	}
	return retMap
}

func (pe *PageElement) GetElementName(ctx PageContext) string {
	return pe.ElementName
}

func (pe *PageElement) GetKind() ElementTagKind {
	return pe.Kind
}
func (pe *PageElement) GetInnerText(ctx PageContext) string {
	return pe.InnerText
}

func RenderPageElement(ctx PageContext, e RenderableElement) (string, error) {
	attr, err := e.GetAttributes(ctx)
	if err != nil {
		return "", err
	}
	switch e.GetKind() {
	case ElementTag_Closing:
		return fmt.Sprintf(`<%s%s>%s</%s>`,
			e.GetElementName(ctx),
			MapToAttributes(attr),
			e.GetInnerText(ctx),
			e.GetElementName(ctx)), nil
	case ElementTag_NoClose:
		return fmt.Sprintf(`<%s%s>`, e.GetElementName(ctx), MapToAttributes(attr)), nil
	case ElementTag_SelfClose:
		return fmt.Sprintf(`<%s%s />`, e.GetElementName(ctx), MapToAttributes(attr)), nil
	default:
		return "", fmt.Errorf("unsupported element kind: %v", e.GetKind())
	}

}

func (pe *PageElement) AddAttribute(key string, val interface{}) *PageElement {
	if pe.Attributes == nil {
		pe.Attributes = map[string]interface{}{}
	}
	pe.Attributes[key] = val
	return pe
}

type PageHead struct {
	Links []*LinkElement
}

type LinkElement struct {
	*PageElement
	LinkType string
	Href     interface{}
}

func (le *LinkElement) GetAttributes(ctx PageContext) (map[string]interface{}, error) {
	m, err := le.PageElement.GetAttributes(ctx)
	if err != nil {
		return nil, err
	}
	url, err := ctx.ResolveUrl(le.Href)
	if err != nil {
		return nil, err
	}
	m["rel"] = le.LinkType
	m["href"] = url
	return m, nil
}

type LinkKind string

const (
	Stylesheet = LinkKind("stylesheet")
)

func (pm *PageHead) AddMeta(vals map[string]interface{}) *PageElement {
	pe := &PageElement{
		AttibutingElement: &AttibutingElement{
			Attributes: vals,
		},
		ElementName: "meta",
	}
	return pe
}

func (pm *PageHead) AddLink(kind LinkKind, source interface{}) *LinkElement {
	l := &LinkElement{
		PageElement: &PageElement{
			Kind:        ElementTag_NoClose,
			ElementName: "link",
		},
		LinkType: string(kind),
		Href:     source,
	}
	pm.Links = append(pm.Links, l)
	return l
}

func (pm *PageHead) Write(ctx PageContext, w io.Writer) error {
	sb := strings.Builder{}
	{
		sb.WriteString("<head>")
		for _, l := range pm.Links {
			html, err := RenderPageElement(ctx, l)
			if err != nil {
				logger.WithError(err).Errorf("failed to render %v", l.LinkType)
				continue
			}

			sb.WriteString(html)

		}
		sb.WriteString("</head>")
	}
	w.Write([]byte(sb.String()))
	return nil
}
