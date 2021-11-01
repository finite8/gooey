package register

// Registerer provides functions for a page or component to access and perform required functions as part of registration
type Registerer interface {
	// Allows a private sub-page to be registered. This cannot be found via menus
	RegisterPrivateSubPage(id string, newPage Page)
}

type registererContext struct {
	currentPage *registeredPageInfo
}

func newRegistererContext(currentPage *registeredPageInfo) Registerer {
	return &registererContext{
		currentPage: currentPage,
	}
}

func (rc *registererContext) RegisterPrivateSubPage(id string, newPage Page) {
	id = rc.currentPage.id + "-" + id
	info := &registeredPageInfo{
		parentid: rc.currentPage.id,
		id:       id,
		page:     newPage,
		private:  true,
	}
	rc.currentPage.children[id] = info
	globalregister.queued[id] = info
	globalregister.registered[id] = info
	globalregister.pageRegister[newPage] = info
}
