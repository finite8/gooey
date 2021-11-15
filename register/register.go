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
		fileSystems:  map[string]http.FileSystem{},
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
	DefaultLayout  PageLayout
)

type webregister struct {
	root         *registeredPageInfo
	registered   map[string]*registeredPageInfo
	queued       map[string]*registeredPageInfo
	pageRegister map[Page]*registeredPageInfo
	mux          sync.Mutex
	layout       PageLayout
	fileSystems  map[string]http.FileSystem
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
		SetLayout(DefaultLayout)
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
	for k, fs := range globalregister.fileSystems {
		http.Handle("/"+k+"/", http.FileServer(fs))
	}
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
	logger.Infof("handling path %s", basePath)
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
		logger.Infof("%T set as layout", l)
	}
	globalregister.layout = l
}

func RegisterFileSystem(name string, fs http.FileSystem) error {
	globalregister.mux.Lock()
	defer globalregister.mux.Unlock()
	if ex, ok := globalregister.fileSystems[name]; ok {
		logger.Warnf("replacing filesystem %s of %T with new filesystem of %T", ex, ex, fs)
	} else {
		logger.Infof("%T set as filesystem '%s'", fs, name)
	}
	globalregister.fileSystems[name] = fs
	return nil
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

func (wr *webregister) getSiteStructure() PageStructure {
	return createpageStructureData(globalregister.root.page)
}

type pageStructureData struct {
	page Page
}

func createpageStructureData(p Page) *pageStructureData {
	return &pageStructureData{
		page: p,
	}
}

func (psd *pageStructureData) Page() Page {
	return psd.page
}
func (psd *pageStructureData) Title() string {
	return psd.page.Name()
}
func (psd *pageStructureData) Children() []PageStructure {
	inf := globalregister.FindPage(psd.page)
	if inf != nil {
		var retArr []PageStructure
		for _, c := range inf.children {
			retArr = append(retArr, createpageStructureData(c.page))
		}
		return retArr
	}
	return nil
}

func (wr *webregister) globalHandler(w http.ResponseWriter, r *http.Request, page Page) {
	ctx := newPageContext(r)
	b := getNewBehaviour(wr.getNewMeta(ctx))
	for _, o := range []interface{}{wr.layout, page} {
		if bc, ok := o.(PageBehaviour); ok {
			b = bc.QueryBehaviour(ctx, b)
		}
	}

	if b.renderHTML {
		// we need to render our HTML stuff
		w.Write([]byte("<html>"))
		b.pageMeta.Write(ctx, w)
	}

	if (b.renderHTML && b.renderLayout) && wr.layout != nil {
		w.Write([]byte("<body>"))
		wr.layout.Render(ctx, w, r, page.Handler)
		w.Write([]byte("</body>"))
	} else {
		page.Handler(ctx, w, r)
	}

	if b.renderHTML {
		// we need to render our HTML stuff
		w.Write([]byte("</html>"))
	}
}

func (wr *webregister) getNewMeta(ctx PageContext) *PageHead {
	return &PageHead{}
}

type Behaviour struct {
	renderLayout bool
	renderHTML   bool
	pageMeta     *PageHead
}

func getNewBehaviour(meta *PageHead) Behaviour {
	return Behaviour{
		renderLayout: true,
		renderHTML:   true,
		pageMeta:     meta,
	}
}

func (b Behaviour) WithRenderLayout(v bool) Behaviour {
	b.renderLayout = v
	return b
}

// WithRenderHTML if disabled, will not render any layouts OR HTML elements. Page response is entirely down to the implementation
func (b Behaviour) WithRenderHTML(v bool) Behaviour {
	b.renderHTML = v
	return b
}

func (b Behaviour) WithMetaHandling(f func(m *PageHead) *PageHead) Behaviour {
	b.pageMeta = f(b.pageMeta)
	return b
}