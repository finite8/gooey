package register

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	logrus "github.com/sirupsen/logrus"
)

// this will allow packages to register themselves with this component. Their UI's will be added into the stack
func init() {
	globalregister = &webregister{
		registered:   map[string]*registeredPageInfo{},
		queued:       map[string]*registeredPageInfo{},
		pageRegister: map[Page]*registeredPageInfo{},
	}
	logger = logrus.New()

}

const (
	RootPageId   = "root"
	SigninPageId = "signin"
)

var (
	globalregister *webregister
	logger         *logrus.Logger
)

type webregister struct {
	root         *registeredPageInfo
	registered   map[string]*registeredPageInfo
	queued       map[string]*registeredPageInfo
	pageRegister map[Page]*registeredPageInfo
	mux          sync.Mutex
	layout       PageLayout
}

func (wr *webregister) FindPage(page Page) *registeredPageInfo {
	return wr.pageRegister[page]
}

func (wr *webregister) RegisterPrivateSubPage(id string, page Page) {

}

type registeredPageInfo struct {
	parentid interface{}
	path     string
	id       string
	page     Page
	children map[string]*registeredPageInfo
	private  bool // if private, the register will not report this page as part of any kind of menu or lookup request
}

func precompileCheck() {
	if globalregister.root == nil {
		logger.Warn("msg", "No Rootpage was provided. Using default Root page")
		RegisterPage(nil, RootPageId, &defaultPage{})
	}
	if globalregister.layout == nil {
		logger.Warn("msg", "No Layout was provided. Using default layout")
		finish this bit
	}
}

// Compile the pages into the required hierachy
func Compile() error {
	precompileCheck()

	globalregister.mux.Lock()
	defer globalregister.mux.Unlock()
	// we are gonna be crude here. We will simply keep looping through the queued items, assigning what we can as we go through. If after a loop there are no additional queued items, we are done
	for {
		itemsCompiled := 0
		for k, v_iter := range globalregister.queued {
			log := logrus.WithField("Id", v_iter.id).WithField("Name", v_iter.page.Name())
			v := v_iter
			var foundParent *registeredPageInfo
			if v.parentid != nil {
				switch tv := v.parentid.(type) {
				case string:
					parentId := tv
					foundParent = globalregister.registered[parentId]
				case Page:
					//var ok bool
					foundParent = globalregister.pageRegister[tv]

				}
			}

			if foundParent != nil {
				foundParent.children[v.id] = v
				log.WithField("ParentId", foundParent.id).Info("Page hierachy established")
			}
			globalregister.registered[v.id] = v
			delete(globalregister.queued, k)
			itemsCompiled++

		}
		if itemsCompiled == 0 {
			break
		}
	}
	return nil
}

func RegisterHandlers() {
	rootPage := globalregister.root

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		globalregister.globalHandler(rw, r, rootPage.page)
	})
	rootPage.path = "/"
	for _, v := range rootPage.children {
		registerHandlersAtPath(rootPage.page, "/", v)
	}

}

func registerHandlersAtPath(parent Page, path string, pageInfo *registeredPageInfo) {
	basePath := strings.ReplaceAll(path+"/"+pageInfo.id, "//", "/")
	http.HandleFunc(basePath, func(rw http.ResponseWriter, r *http.Request) {
		globalregister.globalHandler(rw, r, pageInfo.page)
	})
	pageInfo.path = basePath
	if compl, ok := pageInfo.page.(ComplexPage); ok {
		ctx := newRegistererContext(pageInfo)
		compl.OnHandlerAdded(ctx)
	}
	for _, v := range pageInfo.children {
		registerHandlersAtPath(parent, basePath, v)
	}
}

func SetLayout(l PageLayout) {
	if globalregister.layout != nil {
		logger.Warnf("replacing layout of %t with new layout of %t", globalregister.layout, l)
	} else {
		logger.Infof("%t set as layout", l)
	}
	globalregister.layout = l
}

// RegisterPage adds a renderable page into the system.
// parent: either a resolvable ID or the actual page if available. If nil, it is put in a placeholder location for later referencing. parents may not exist yet at time of this being called
func RegisterPage(parent interface{}, id string, page Page) error {
	globalregister.mux.Lock()
	defer globalregister.mux.Unlock()
	id = strings.TrimSpace(strings.ToLower(id))
	pageInfo := &registeredPageInfo{
		parentid: parent,
		id:       id,
		page:     page,
		children: make(map[string]*registeredPageInfo),
	}
	log := logrus.WithField("Id", id)
	if id == RootPageId {
		if globalregister.root != nil {
			// but the root has already been defined. Lets die
			err := errors.New("invalid root page registration: root has already been registered")
			log.Fatal(err)
		}
		// we have a root page being registered
		globalregister.root = pageInfo
		globalregister.registered[RootPageId] = pageInfo
		log.Info("Root page registered")
	} else {
		// this is not a root page. Coolies.
		// we are not going to try and shuffle and organize things yet. We will just queue it all up first.
		globalregister.queued[id] = pageInfo
		if pageInfo.parentid != nil {
			// check to make sure it is a valid type
			switch pageInfo.parentid.(type) {
			case string:
			case Page:
			default:
				err := fmt.Errorf("invalid page registration: parent type %t not allowed", pageInfo.parentid)
				log.Fatal(err)
			}
		}
		log.Info("Page registered")
	}
	globalregister.pageRegister[page] = pageInfo
	return nil

}

func (wr *webregister) globalHandler(w http.ResponseWriter, r *http.Request, page Page) {
	ctx := newPageContext(r)
	b := getNewBehaviour()
	if bc, ok := page.(PageBehaviour); ok {
		b = bc.QueryBehaviour(ctx, b)
	}
	if b.renderLayout && wr.layout != nil {
		wr.layout.RenderLeading(ctx, w)
	}
	page.Handler(ctx, w, r)
	if b.renderLayout && wr.layout != nil {
		wr.layout.RenderTrailing(ctx, w)
	}
}

type Behaviour struct {
	renderLayout bool
}

func getNewBehaviour() Behaviour {
	return Behaviour{
		renderLayout: true,
	}
}

func (b Behaviour) WithRenderLayout(v bool) Behaviour {
	b.renderLayout = v
	return b
}
