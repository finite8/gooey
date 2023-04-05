package register

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// PageContext provides functionality that a component or page might need to perform its functions
type PageContext interface {
	context.Context
	GetPageUrl(p Page) *url.URL
	// Context data refers to any additional parameters or modifiers that have changed how the page has been called (i.e: query string in URL)
	GetContextData() map[string][]string
	UnmarshallData(interface{})
	ResolveUrl(i interface{}) (*url.URL, error)
	SiteRoot() PageStructure
	Resolve(i interface{}, rk ResolutionKind) string
	// Gets the Request cache: a key-value store that will persist for the life of the request
	RequestCache() Cache
}

var _ PageContext = (*pageContext)(nil)

type pageContext struct {
	context.Context
	request  *http.Request
	reqCache Cache
}

func newPageContext(ctx context.Context, r *http.Request) *pageContext {
	return &pageContext{
		Context:  ctx,
		request:  r,
		reqCache: newMemoryCache(),
	}
}

func (pctx *pageContext) RequestCache() Cache {
	return pctx.reqCache
}

func (pctx *pageContext) SiteRoot() PageStructure {
	return globalregister.getSiteStructure()
}

func (pctx *pageContext) GetContextData() map[string][]string {

	return pctx.request.URL.Query()
}
func (pctx *pageContext) UnmarshallData(v interface{}) {
	err := pctx.request.ParseForm()
	if err != nil {
		logrus.Error(err)
		return
	}
	// valMap := make(map[string]interface{})
	// for k, v := range pctx.request.Form() {

	// }
	// first, we need to simplify our structure as required.

}

func (pctx *pageContext) buildUrlFromRoot(path string) string {
	return fmt.Sprintf("%s://%s", getSchemeFromProto(pctx.request.Proto), pctx.request.Host) + path
}

func (pctx *pageContext) GetPageUrl(p Page) *url.URL {
	info := globalregister.FindPage(p)
	if info == nil {
		panic("failed to resolve required page")
	}
	basePath := pctx.buildUrlFromRoot(info.path)

	u, err := url.Parse(basePath)
	if err != nil {
		panic(err)
	}
	return u
}

func getSchemeFromProto(proto string) string {
	parts := strings.Split(proto, "/")
	switch {
	case strings.EqualFold(parts[0], "http"):
		return "http"
	}
	return "http"
}

func (pctx *pageContext) Resolve(i interface{}, rk ResolutionKind) string {
	switch it := i.(type) {
	case Resolvable:
		return it.Resolve(pctx, rk)
	case Page:
		switch rk {
		case Resolution_CSSClass:
			return "" // a page cannot be used here
		}
	case string:
		return it
	case *string:
		if it != nil {
			return *it
		}
	}
	return ""
}

func (pctx *pageContext) ResolveUrl(i interface{}) (*url.URL, error) {
	if p, ok := i.(Page); ok {
		return pctx.GetPageUrl(p), nil
	}
	// so, it wasn't the obvious, so lets see what else we have
	switch v := i.(type) {
	case GOOEYFile:
		// this is a gooey file
		return url.Parse(pctx.buildUrlFromRoot(v.FullPath()))
	case string:
		// in this case, we assume the path is relative
		return pctx.request.URL.Parse(v)
	case *url.URL:
		return pctx.request.URL.ResolveReference(v), nil
	default:
		return nil, fmt.Errorf("cannot use %t to resolve a url", i)
	}
}
