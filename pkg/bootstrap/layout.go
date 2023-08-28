package bootstrap

import (
	"net/http"

	_ "embed"

	"github.com/finite8/gooey/core"
	"github.com/finite8/gooey/register"
)

//go:embed css/bootstrapbind.css
var bootstrapBindCSS []byte

//go:embed src/dist/css/bootstrap.min.css
var bootstrapCSS []byte

//go:embed src/dist/js/bootstrap.min.js
var bootstrapJS []byte

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

	bl.bscFile = core.CoreFS.SetFileBytes("bootstrapCompat.scss", bootstrapBindCSS)
	bl.bsFile = core.CoreFS.SetFileBytes("bootstrap.min.css", bootstrapCSS)
	bl.jsFile = core.CoreFS.SetFileBytes("bootstrap.bundle.min.js", bootstrapJS)
	register.SetLayout(bl)

}

type BootstrapLayout struct {
	*core.CommonLayout
	bscFile register.GOOEYFile
	jsFile  register.GOOEYFile
	bsFile  register.GOOEYFile
}

func (bl BootstrapLayout) String() string {
	return "BootstrapLayout"
}

func (bl *BootstrapLayout) QueryBehaviour(ctx register.PageContext, b register.Behaviour) register.Behaviour {
	b.WithMetaHandling(func(m *register.PageHead) *register.PageHead {
		m.AddLink(register.Stylesheet, bl.bsFile)
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

	// jqueryElem, _ := register.RenderPageElement(ctx, &register.PageElement{
	// 	ElementName: "script",
	// 	Kind:        register.ElementTag_Closing,
	// 	AttibutingElement: &register.AttibutingElement{
	// 		Attributes: map[string]interface{}{
	// 			"src":         "https://code.jquery.com/jquery-3.2.1.slim.min.js",
	// 			"integrity":   "sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN",
	// 			"crossorigin": "anonymous",
	// 		},
	// 	},
	// })
	// popperElem, _ := register.RenderPageElement(ctx, &register.PageElement{
	// 	ElementName: "script",
	// 	Kind:        register.ElementTag_Closing,
	// 	AttibutingElement: &register.AttibutingElement{
	// 		Attributes: map[string]interface{}{
	// 			"src":         "https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js",
	// 			"integrity":   "sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q",
	// 			"crossorigin": "anonymous",
	// 		},
	// 	},
	// })
	// bootstrapElem, _ := register.RenderPageElement(ctx, &register.PageElement{
	// 	ElementName: "script",
	// 	Kind:        register.ElementTag_Closing,
	// 	AttibutingElement: &register.AttibutingElement{
	// 		Attributes: map[string]interface{}{
	// 			"src":         "https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js",
	// 			"integrity":   "sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl",
	// 			"crossorigin": "anonymous",
	// 		},
	// 	},
	// })
	bootstrapElem, _ := register.RenderPageElement(ctx, &register.PageElement{
		ElementName: "script",
		Kind:        register.ElementTag_Closing,
		AttibutingElement: &register.AttibutingElement{
			Attributes: map[string]interface{}{
				"src": bl.jsFile,
				// "integrity":   "sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q",
				// "crossorigin": "anonymous",
			},
		},
	})
	// w.Write([]byte(jqueryElem + "\n"))
	//w.Write([]byte(popperElem + "\n"))
	w.Write([]byte(bootstrapElem + "\n"))

}
