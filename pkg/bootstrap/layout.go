//go:generate go run ../staticerize/. -in ./css/bootstrapbind.css -out bootstrapCSS.go -pkg bootstrap -var bootstrapCSS

package bootstrap

import (
	"net/http"

	"github.com/ntaylor-barnett/gooey/core"
	"github.com/ntaylor-barnett/gooey/register"
)

const gooeyBootstrapBind = `
@use 'https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css';
.GOOEY_nav {
	@extend .navbar;
	@extend .navbar-expand-lg;
	@extend .navbar-light; 
	@extend .bg-light;
}
`

func init() {
	bl := &BootstrapLayout{
		CommonLayout: core.NewCommonLayout(),
	}
	navList := core.NewListComponent(func(pc register.PageContext) (interface{}, error) {
		navPages := append([]register.PageStructure{pc.SiteRoot()}, pc.SiteRoot().Children()...)
		return navPages, nil
	})
	navList.Style = core.NewClassStyling("navbar navbar-expand-lg navbar-light bg-light")
	bl.Nav = navList

	bl.bscFile = core.CoreFS.SetFileString("bootstrapCompat.scss", bootstrapCSS)
	register.SetLayout(bl)

}

type BootstrapLayout struct {
	*core.CommonLayout
	bscFile register.GOOEYFile
}

func (bl BootstrapLayout) String() string {
	return "BootstrapLayout"
}

func (bl *BootstrapLayout) QueryBehaviour(ctx register.PageContext, b register.Behaviour) register.Behaviour {
	b.WithMetaHandling(func(m *register.PageHead) *register.PageHead {
		m.AddLink(register.Stylesheet, "https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css")
		m.AddLink(register.Stylesheet, bl.bscFile)
		m.AddMeta(map[string]interface{}{
			"charset": "utf-8",
		})
		m.AddMeta(map[string]interface{}{
			"name":    "viewport",
			"content": "width=device-width, initial-scale=1, shrink-to-fit=no",
		})
		return m
	})

	return bl.CommonLayout.QueryBehaviour(ctx, b)
}

func (bl *BootstrapLayout) Render(ctx register.PageContext, w http.ResponseWriter, r *http.Request, pageRenderer func(ctx register.PageContext, w http.ResponseWriter, r *http.Request)) {

	bl.CommonLayout.Render(ctx, w, r, pageRenderer)
	// now we need to add our own script stuff so it renders right

	jqueryElem, _ := register.RenderPageElement(ctx, &register.PageElement{
		ElementName: "script",
		Kind:        register.ElementTag_Closing,
		AttibutingElement: &register.AttibutingElement{
			Attributes: map[string]interface{}{
				"src":         "https://code.jquery.com/jquery-3.2.1.slim.min.js",
				"integrity":   "sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN",
				"crossorigin": "anonymous",
			},
		},
	})
	popperElem, _ := register.RenderPageElement(ctx, &register.PageElement{
		ElementName: "script",
		Kind:        register.ElementTag_Closing,
		AttibutingElement: &register.AttibutingElement{
			Attributes: map[string]interface{}{
				"src":         "https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js",
				"integrity":   "sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q",
				"crossorigin": "anonymous",
			},
		},
	})
	bootstrapElem, _ := register.RenderPageElement(ctx, &register.PageElement{
		ElementName: "script",
		Kind:        register.ElementTag_Closing,
		AttibutingElement: &register.AttibutingElement{
			Attributes: map[string]interface{}{
				"src":         "https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js",
				"integrity":   "sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl",
				"crossorigin": "anonymous",
			},
		},
	})
	w.Write([]byte(jqueryElem + "\n"))
	w.Write([]byte(popperElem + "\n"))
	w.Write([]byte(bootstrapElem + "\n"))

}
