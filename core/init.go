package core

import (
	"sort"

	"github.com/ntaylor-barnett/gooey/register"
)

func init() {
	cl := NewCommonLayout()
	cl.Nav = NewListComponent(func(pc register.PageContext) (interface{}, error) {
		navPages := append(register.PageStructureCollection{pc.SiteRoot()}, pc.SiteRoot().Children()...)
		sort.Sort(navPages)
		return navPages, nil
	})
	register.DefaultLayout = cl
	CoreFS = *register.NewVirtualFS("GoopCoreAssets")
	register.RegisterFileSystem("GoopCoreAssets", CoreFS)
}

var CoreFS register.VirtualFS
