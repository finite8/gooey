package core

import (
	"sort"

	"github.com/ntaylor-barnett/gooey/register"
)

func init() {
	register.DefaultLayout = createBaseLayout()
	register.CorePages().ErrorPage = createBaseErrorPage()
	CoreFS = *register.NewVirtualFS("GoopCoreAssets")
	register.RegisterFileSystem("GoopCoreAssets", CoreFS)
}

func createBaseLayout() *CommonLayout {
	cl := NewCommonLayout()
	cl.Nav = NewListComponent(func(pc register.PageContext) (interface{}, error) {
		navPages := append(register.PageStructureCollection{pc.SiteRoot()}, pc.SiteRoot().Children()...)
		sort.Sort(navPages)
		return navPages, nil
	})
	return cl
}

func createBaseErrorPage() Page {
	ep := &ContainerPage{}
	ep = ep.WithName("error").WithComponent(NewContainerComponent().WithStyle(NewCardRenderer().WithWidthPortion(Portion75)).WithComponent(NewTextPrimitve("oops")))
	return ep
}

var CoreFS register.VirtualFS
