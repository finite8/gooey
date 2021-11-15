package core

import "github.com/ntaylor-barnett/gooey/register"

func init() {
	cl := NewCommonLayout()
	cl.Nav = NewListComponent(func(pc register.PageContext) (interface{}, error) {
		navPages := append([]register.PageStructure{pc.SiteRoot()}, pc.SiteRoot().Children()...)
		return navPages, nil
	})
	register.DefaultLayout = cl
	CoreFS = *register.NewVirtualFS("GoopCoreAssets")
	register.RegisterFileSystem("GoopCoreAssets", CoreFS)
}

var CoreFS register.VirtualFS
